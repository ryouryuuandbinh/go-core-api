package utils

import (
	"go-core-api/pkg/custom_error"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext trích xuất an toàn user_id từ Gin Context (do Middleware truyền vào)
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		return 0, custom_error.ErrUnauthorized
	}

	userIDVal, ok := userIDFloat.(uint)
	if !ok {
		return 0, custom_error.ErrUnauthorized
	}

	return userIDVal, nil
}
