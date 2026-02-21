// Package database contains the logic for database connections and migrations
package database

import (
	"go-core-api/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB Biến toàn cục để các package khác có thể gọi
var DB *gorm.DB

// ConnectDB khởi tạo kết nối tới PostgreSQL, dsn (Data Source Name): Chuỗi chứa các thông tin kết nối(host, user, pass, dbname)
func ConnectDB(dsn string) {
	var err error

	// 1. Mở kết nối qua gorm
	// gorm.Config: cấu hình hoạt động của gorm
	// logger.Default.LogMode(logger.Info): Quan trọng! Nó sẽ in mọi câu lệnh SQL ra terminal
	// -> Giúp bạn debug xem GORM đang "dịch" code Go sang SQL như thế nào.
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		// Nếu kết nối thất bại, dừng chương trình ngay lập tức (Panic/Fatal)
		logger.Fatal("❌ Lỗi kết nối Database: %v", zap.Error(err))
	}

	// 2. Cấu hình Connection Pool (Hồ chứa kết nối)
	// Lấy đối tượng sql.DB nguyên thủy từ GORM để cấu hình sâu hơn
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Fatal("❌ Lỗi kết nối Database Pool: %v", zap.Error(err))
	}

	// Thuật toán connection Pool
	// SetMaxIdleConns: Số kết nối nhãn rỗi tối đa giữ lại trong pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns: Số kết nối tối đa mở cùng lúc. Vượt qua số này, resquest phải đợi.
	sqlDB.SetMaxOpenConns(100)
	// SetMaxLifeTime: Thời gian sống tối đa của 1 kết nối để tránh lỗi timeout từ phía DB.
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("✅ Kết nối Database thành công!")
}
