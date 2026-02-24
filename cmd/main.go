package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-core-api/internal/middlewares"
	"go-core-api/internal/models"
	"go-core-api/internal/server" // Bổ sung import package server mới tạo
	"go-core-api/pkg/config"
	"go-core-api/pkg/database"
	"go-core-api/pkg/logger"
	"go-core-api/pkg/mailer"
	"go-core-api/pkg/utils"

	"go.uber.org/zap"
)

func main() {

	// 1. INFRASTRUCTURE
	logger.InitLogger()
	defer logger.Log.Sync()

	config.LoadConfig()
	cfg := config.AppConfig

	database.ConnectDB(cfg.Database.DSN)
	database.DB.AutoMigrate(&models.User{})

	mailService := mailer.NewMailer(
		cfg.Mailer.Host, cfg.Mailer.Port,
		cfg.Mailer.User, cfg.Mailer.Password, cfg.Mailer.From,
	)

	// 2. (BACKGROUND WORKERS)
	ctxWorker, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	utils.InitWorkerPool(ctxWorker, 20)
	go middlewares.InitRateLimiterCleanup(ctxWorker)

	// 3. DEPENDENCY INJECTION & ROUTER
	r := server.SetupDependenciesAndRouter(database.DB, cfg, mailService)

	// 4. SERVER & GRACEFUL SHUTDOWN
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		logger.Info("🚀 Server đang chạy tại: " + cfg.Server.Domain)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Lỗi khởi chạy server", zap.Error(err))
		}
	}()

	// Chờ tín hiệu tắt từ hệ điều hành (Ctrl+C, Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Đang tắt Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Lỗi khi tắt Server", zap.Error(err))
	}

	logger.Info("Đang chờ các tác vụ nền hoàn tất...")
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer waitCancel()

	c := make(chan struct{})
	go func() {
		defer close(c)
		utils.WorkerGroup.Wait()
	}()

	select {
	case <-c:
		logger.Info("✅ Tất cả worker đã hoàn tất an toàn.")
	case <-waitCtx.Done():
		logger.Error("❌ Timeout: Ép buộc tắt tiến trình do worker bị treo.")
	}

	logger.Info("Server đã tắt hoàn toàn.")
}
