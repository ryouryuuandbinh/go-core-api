package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log là biến toàn cục để các package khác gọi: logger.log.Info("...")
var Log *zap.Logger

// InitLogger khởi tạo hệ thống log
func InitLogger() {
	// 1. Cấu hình định dạng log (Encoder)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Format thời gian dễ đọc: 2026-02-21T...
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Màu sắc cho level (INFO xanh, ERROR đỏ)

	// Dùng ConsoleEncoder để in ra chữ dễ nhìn (Nếu làm API Production thực tế hay dùng JSONEncoder)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 2. Cấu hình nơi lưu file log (Dùng Lumberjack để cắt file)
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/api.log", // Đường dẫn file log
		MaxSize:    10,             // Dung lượng tối đa 1 file (Megabytes)
		MaxAge:     30,             // Giữ tối đa 30 ngày
		MaxBackups: 5,              // Giữ lại tối đa 5 file cũ
		Compress:   true,           // Nén file cũ (zip)
	})

	// 3. Cấu hình lõi (Core) của Logger: ghi ra cả Terminal (os.Stdout) và File
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),       // In ra Terminal
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileWriter, zapcore.InfoLevel), // Lưu file dạng JSON
	)

	// 4. Khởi tạo Logger
	Log = zap.New(core, zap.AddCaller()) // AddCaller để biết log được gọi từ file nào, dòng số mấy
}

// Các hàm bọc (Wrapper) để gọi cho nhanh
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}
