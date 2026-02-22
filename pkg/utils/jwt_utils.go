package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext trích xuất an toàn user_id từ Gin Context (do Middleware truyền vào)
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("không tìm thấy thông tin xác thực")
	}

	userIDVal, ok := userIDFloat.(float64)
	if !ok {
		return 0, errors.New("token sai định dạng")
	}

	return uint(userIDVal), nil
}
