package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-core-api/internal/handlers"
	"go-core-api/internal/middlewares"
	"go-core-api/internal/models"
	"go-core-api/internal/repositories"
	"go-core-api/internal/services"
	"go-core-api/pkg/config"
	"go-core-api/pkg/database"
	"go-core-api/pkg/mailer"

	"github.com/gin-gonic/gin"
)

func main() {
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
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, mailService, cfg.JWT.Secret)
	userHandler := handlers.NewUserHandler(userService)
	uploadHandler := handlers.NewUploadHandler()

	// 3.Khởi tạo Router gin
	r := gin.Default()

	// --- QUAN TRỌNG: Cấu hình phục vụ file tĩnh ---
	// Khi user truy cập http://domain/upload/xxx.jpg -> nó sẽ tìm file trong thư mục "./uploads"
	r.Static("uploads", "./uploads")

	// 4. Khai báo API Endpoints
	v1 := r.Group("/api/v1")
	{
		// API không cần Auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh-token", authHandler.RefreshToken)
		}

		// API cần Auth & Test phân quyền
		protected := v1.Group("/admin")
		protected.Use(middlewares.RequireAuth(cfg.JWT.Secret), middlewares.RequireRole(models.RoleAdmin))
		{
			protected.GET("/dashboard", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				c.JSON(200, gin.H{"message": "Chào mừng Admin!", "your_id": userID})
			})
		}

		// API Upload (Cần đăng nhập mới được upload image)
		upload := v1.Group("/upload")
		upload.Use(middlewares.RequireAuth(cfg.JWT.Secret))
		{
			upload.POST("/image", uploadHandler.UploadImage)
		}

		// API Users (Chỉ Admin mới xem được danh sách)
		userRouters := v1.Group("/users")
		userRouters.Use(middlewares.RequireAuth(cfg.JWT.Secret))
		{
			// Chỉ Admin mới xem được danh sách

			// User nào cũng tự đổi password của mình được
			userRouters.PUT("/me/password", userHandler.ChangePassword)

			// Lấy thông tin cá nhân
			userRouters.GET("/me", userHandler.GetMe)

			// Cập nhật thông tin cá nhân
			userRouters.PUT("/me", userHandler.UpdateProfile)

			// --- API CỦA ADMIN ---
			adminUserRouters := userRouters.Group("")
			adminUserRouters.Use(middlewares.RequireRole(models.RoleAdmin))
			{
				adminUserRouters.GET("", userHandler.GetList)
				adminUserRouters.GET("/:id", userHandler.GetUser)
				adminUserRouters.PUT("/:id", userHandler.AdminUpdateUser)
				adminUserRouters.DELETE("/:id", userHandler.DeleteUser)
			}
		}
	}

	// Chạy Server bằng config port
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// Chạy server trong 1 goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi khởi chạy server: %s\n", err)
		}
	}()

	// Chờ tín hiệu tắt (Ctrl+C hoặc Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Đang tắt Server...")

	// Cho server 5 giây để xử lý nốt các request đang dang dở
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Lỗi khi tắt Server:", err)
	}

	log.Println("Server đã tắt an toàn.")
}
