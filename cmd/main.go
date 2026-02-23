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
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret)
	userService := services.NewUserService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, mailService, cfg.JWT.Secret)
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
			logger.Fatal("Lỗi khởi chạy server", zap.Error(err)) // Đổi ở đây
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
	utils.WorkerGroup.Wait()

	logger.Info("Server đã tắt an toàn.")
}
