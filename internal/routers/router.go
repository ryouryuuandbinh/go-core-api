package routers

import (
	"time"

	"go-core-api/internal/handlers"
	"go-core-api/internal/middlewares"
	"go-core-api/internal/models"
	"go-core-api/internal/repositories"
	"go-core-api/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	uploadHandler *handlers.UploadHandler,
	userRepo repositories.UserRepository,
) *gin.Engine {
	r := gin.New()
	cfg := config.AppConfig

	r.Use(middlewares.ZapLogger(), gin.Recovery())

	// SỬA LỖI CORS: Cấu hình chuẩn W3C
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.Server.Domain}, // Thay "*" bằng domain Frontend thực tế
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/uploads", "./uploads")

	// Áp dụng giới hạn ram mặc định cho file tải lên ở level Router (8MB bộ nhớ RAM, phần thừa ghi ra temp disk)
	r.MaxMultipartMemory = 8 << 20

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.Use(middlewares.RateLimitMiddleware())
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh-token", authHandler.RefreshToken)
			auth.POST("/logout", middlewares.RequireAuth(cfg.JWT.Secret, userRepo), authHandler.Logout)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		protected := v1.Group("/admin")
		protected.Use(middlewares.RequireAuth(cfg.JWT.Secret, userRepo), middlewares.RequireRole(models.RoleAdmin))
		{
			protected.GET("/dashboard", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				c.JSON(200, gin.H{"message": "Chào mừng Admin!", "your_id": userID})
			})
		}

		upload := v1.Group("/upload")
		upload.Use(middlewares.RequireAuth(cfg.JWT.Secret, userRepo))
		{
			upload.POST("/image", uploadHandler.UploadImage)
		}

		userRouters := v1.Group("/users")
		userRouters.Use(middlewares.RequireAuth(cfg.JWT.Secret, userRepo))
		{
			userRouters.PUT("/me/password", userHandler.ChangePassword)
			userRouters.GET("/me", userHandler.GetMe)
			userRouters.PUT("/me", userHandler.UpdateProfile)

			adminUserRouters := userRouters.Group("")
			adminUserRouters.Use(middlewares.RequireRole(models.RoleAdmin))
			{
				adminUserRouters.GET("", userHandler.GetList)
				adminUserRouters.GET("/:id", userHandler.GetUser)
				adminUserRouters.PUT("/:id", userHandler.AdminUpdateUser)
				adminUserRouters.DELETE("/:id", userHandler.DeleteUser)
				adminUserRouters.DELETE("/:id/purge", userHandler.PurgeUser)
			}
		}
	}

	return r
}
