package middlewares

import (
	"time"

	"go-core-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger thay thế logger mặc định của Gin
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bắt đầu đo thời gian
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Cho phép request đi tiếp vào Handler xử lý
		c.Next()

		// Sau khi xử lý xong, tính toán thời gian và lấy thông tin
		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Ghi log bằng Zap siêu tốc
		logger.Log.Info("API Request",
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		)
	}
}
