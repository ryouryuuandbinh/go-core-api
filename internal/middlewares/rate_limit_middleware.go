package middlewares

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen int64 // Dùng kiểu int64 để áp dụng các phép toán atomic (không cần lock)
}

// Thay thế map tĩnh và Mutex bằng sync.Map siêu tốc của Go
var visitors sync.Map

func init() {
	go cleanupVisitors()
}

func getLimiter(ip string) *rate.Limiter {
	// Lấy giá trị từ sync.Map
	v, exists := visitors.Load(ip)
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(5), 10)
		newVisitor := &visitor{
			limiter:  limiter,
			lastSeen: time.Now().Unix(),
		}
		// Đảm bảo an toàn khi nhiều goroutine cùng cố gắng thêm 1 IP
		v, _ = visitors.LoadOrStore(ip, newVisitor)
	}

	vis := v.(*visitor)
	// Cập nhật lastSeen không cần dùng khóa (Lock-free)
	atomic.StoreInt64(&vis.lastSeen, time.Now().Unix())
	return vis.limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(3 * time.Minute)
		now := time.Now().Unix()

		visitors.Range(func(key, value interface{}) bool {
			vis := value.(*visitor)
			// Nếu không có tương tác sau 3 phút (180 giây) -> Xóa
			if now-atomic.LoadInt64(&vis.lastSeen) > 180 {
				visitors.Delete(key)
			}
			return true // Tiếp tục vòng lặp Range
		})
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "Bạn đã gửi quá nhiều yêu cầu. Vui lòng thử lại sau.")
			c.Abort()
			return
		}
		c.Next()
	}
}
