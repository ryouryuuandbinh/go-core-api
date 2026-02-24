package server

import (
	"os"

	"go-core-api/internal/handlers"
	"go-core-api/internal/repositories"
	"go-core-api/internal/routers"
	"go-core-api/internal/services"
	"go-core-api/pkg/config"
	"go-core-api/pkg/logger"
	"go-core-api/pkg/mailer"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupDependenciesAndRouter gom toàn bộ logic tiêm phụ thuộc (DI) vào một chỗ
func SetupDependenciesAndRouter(db *gorm.DB, cfg *config.Config, mailService mailer.Mailer) *gin.Engine {
	// 1. Cấu hình môi trường cho Gin
	gin.SetMode(gin.ReleaseMode)

	// Tạo thư mục uploads nếu chưa có (Tránh lỗi vặt khi chạy ở máy mới)
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		logger.Fatal("Không thể tạo thư mục uploads", zap.Error(err))
	}

	// 2. Khởi tạo tầng Repositories (Data Access)
	userRepo := repositories.NewUserRepository(db)

	// 3. Khởi tạo tầng Services (Business Logic)
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret, mailService)
	userService := services.NewUserService(userRepo)

	// 4. Khởi tạo tầng Handlers (HTTP Layer)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	uploadHandler := handlers.NewUploadHandler()

	// 5. Ráp tất cả vào Router và trả về
	return routers.SetupRouter(authHandler, userHandler, uploadHandler, userRepo)
}
