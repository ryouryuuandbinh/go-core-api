package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-core-api/internal/handlers"
	"go-core-api/internal/middlewares"
	"go-core-api/internal/models"
	"go-core-api/internal/repositories"
	"go-core-api/internal/routers"
	"go-core-api/internal/services"
	"go-core-api/pkg/config"
	"go-core-api/pkg/database"
	"go-core-api/pkg/logger"
	"go-core-api/pkg/mailer"
	"go-core-api/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {

	logger.InitLogger()
	// Quét dọn bộ nhớ log trước khi tắt app
	defer logger.Log.Sync()

	// 1. Load config
	config.LoadConfig()
	cfg := config.AppConfig

	// --- [SỬA LỖI GOROUTINE LEAK] ---
	// Khởi tạo một Context bao trùm cho tất cả các Background Worker (như Rate Limiter Cleanup)
	// Khi cancelWorker() được gọi (lúc tắt server), mọi vòng lặp for vô tận nhận ctxWorker này sẽ tự động dừng lại.
	ctxWorker, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	// KHỞI TẠO WORKER POOL CHUẨN MỰC
	utils.InitWorkerPool(ctxWorker, 20)

	// Khởi chạy dọn dẹp Rate Limiter ngầm
	go middlewares.InitRateLimiterCleanup(ctxWorker)
	// --------------------------------

	// 2. Khởi tạo DB dùng config
	database.ConnectDB(cfg.Database.DSN)
	database.DB.AutoMigrate(&models.User{})

	// 3. Khởi tạo Mailer dùng config
	mailService := mailer.NewMailer(
		cfg.Mailer.Host,
		cfg.Mailer.Port,
		cfg.Mailer.User,
		cfg.Mailer.Password,
		cfg.Mailer.From,
	)

	// 4. Dependency Injection (Bơm phụ thuộc từ dưới lên)
	userRepo := repositories.NewUserRepository(database.DB)

	// --- [SỬA LỖI COMPILE ERROR (DI)] ---
	// Bơm mailService vào AuthService, thay vì bơm vào AuthHandler như trước
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret, mailService)
	userService := services.NewUserService(userRepo)

	// AuthHandler giờ đây rất "sạch", chỉ nhận đúng authService
	authHandler := handlers.NewAuthHandler(authService)
	// ------------------------------------

	userHandler := handlers.NewUserHandler(userService)
	uploadHandler := handlers.NewUploadHandler()

	// Tắt log debug của Gin
	gin.SetMode(gin.ReleaseMode)

	// Tạo thư mục uploads nếu chưa có
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		logger.Fatal("Không thể tạo thư mục uploads", zap.Error(err))
	}

	r := routers.SetupRouter(authHandler, userHandler, uploadHandler, userRepo)

	// Chạy Server bằng config port
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// Chạy server trong 1 goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Lỗi khởi chạy server", zap.Error(err))
		}
	}()

	// Chờ tín hiệu tắt (Ctrl+C hoặc Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Đang tắt Server...")

	// Cho server 5 giây để xử lý nốt các request đang dang dở
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Lỗi khi tắt Server", zap.Error(err))
	}

	logger.Info("Đang chờ các tác vụ nền hoàn tất...")
	// Khắc phục lỗi Deadlock bằng Wait Timeout
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer waitCancel()

	// Tạo 1 channel để báo hiệu WaitGroup đã xong
	c := make(chan struct{})
	go func() {
		defer close(c)
		utils.WorkerGroup.Wait()
	}()

	select {
	case <-c:
		logger.Info("Tất cả worker đã hoàn tất an toàn.")
	case <-waitCtx.Done():
		logger.Error("Timeout: Ép buộc tắt tiến trình do worker bị treo.")
	}

	logger.Info("Server đã tắt an toàn.")
}
