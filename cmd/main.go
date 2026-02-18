package main

import (
	"go-core-api/internal/handlers"
	"go-core-api/internal/middlewares"
	"go-core-api/internal/models"
	"go-core-api/internal/repositories"
	"go-core-api/internal/services"
	"go-core-api/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Khởi tạo DB
	dsn := "host=localhost user=postgres password=123456 dbname=core_api port=5432 sslmode=disable"
	jwtSecret := "key-secret"
	database.ConnectDB(dsn)
	database.DB.AutoMigrate(&models.User{}) // Tự động tạo bảng trong database

	// 2. Dependency Injection (Bơm phụ thuộc từ dưới lên)
	userRepo := repositories.NewUserRepository(database.DB)
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, jwtSecret)
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
		}

		// API cần Auth & Test phân quyền
		protected := v1.Group("/admin")
		protected.Use(middlewares.RequireAuth(jwtSecret), middlewares.RequireRole(models.RoleAdmin))
		{
			protected.GET("/dashboard", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				c.JSON(200, gin.H{"message": "Chào mừng Admin!", "your_id": userID})
			})
		}

		// API Upload (Cần đăng nhập mới được upload image)
		upload := v1.Group("/upload")
		upload.Use(middlewares.RequireAuth(jwtSecret))
		{
			upload.POST("/image", uploadHandler.UploadImage)
		}

		// API Users (Chỉ Admin mới xem được danh sách)
		userRouters := v1.Group("/users")
		userRouters.Use(middlewares.RequireAuth(jwtSecret))
		{
			userRouters.GET("", userHandler.GetList)
		}
	}

	// 5. Chạy Server
	r.Run(":8000")
}
