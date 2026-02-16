package response

import (
	"github.com/gin-gonic/gin"
)

// Chuẩn hoá cấu trúc trả về
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // Nếu null sẽ không hiển thị json field này
}

// Hàm Success chuẩn hoá
func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Hàm Error chuẩn hoá
func Error(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, Response{
		Code:    code,
		Message: message,
	})
}
