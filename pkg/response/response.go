package response

import (
	"net/http"

	"go-core-api/pkg/custom_error"

	"github.com/gin-gonic/gin"
)

// Chuẩn hoá cấu trúc trả về
type Response struct {
	Code    int         `json:"code"`
	ErrCode string      `json:"err_code,omitempty"`
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
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*custom_error.AppError); ok {
		c.AbortWithStatusJSON(appErr.HTTPCode, Response{
			Code:    appErr.HTTPCode,
			ErrCode: appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	// 2. Nếu là lỗi văng ra từ hệ thống (VD: Lỗi connect DB, lỗi code crash)
	c.AbortWithStatusJSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		ErrCode: custom_error.ErrInternalServer.Code,
		Message: err.Error(), // Ở production thực tế có thể log ra file thay vì in cho user
	})
}
