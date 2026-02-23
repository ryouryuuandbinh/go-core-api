package middlewares

import (
	"net/http"
	"sync"
	"time"

	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Gói limiter kèm theo thời gian truy cập cuối để tiện dọn rác
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var visitors = make(map[string]*visitor)
var mu sync.Mutex

// init() tự động chạy khi package được nạp, giúp khởi chạy trình dọn rác ngầm
func init() {
	go cleanupVisitors()
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(5), 10)
		visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors quét và xoá các IP rác mỗi 3 phút để chống Memory Leak
func cleanupVisitors() {
	ticker := time.NewTicker(3 * time.Minute)
	for range ticker.C {
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
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
