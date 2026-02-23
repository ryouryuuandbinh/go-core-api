package middlewares

import (
	"net/http"
	"sync"

	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Dùng map để lưu Rate Limiter cho từng IP riêng biệt
var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

// getLimiter trả về limiter cho một IP cụ thể (giới hạn 5 requests/giây, tối đa 10 requests cùng lúc)
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		// rate.Limit(5): 5 token hồi lại mỗi giây
		// 10: Burst size (Cho phép tối đa 10 request dồn dập trong 1 khoảnh khắc)
		limiter = rate.NewLimiter(rate.Limit(5), 10)
		visitors[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware kiểm soát lưu lượng truy cập
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
