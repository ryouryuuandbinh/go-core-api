package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy Token từ Header: Authorization: Bearer <Token>
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			response.Error(c, http.StatusUnauthorized, "Yêu cầu đăng nhập")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "Token sai định dạng")
			return
		}

		tokenString := parts[1]

		// 2. Validate Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("phương thức mã hoá không hợp lệ")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, http.StatusUnauthorized, "Token hết hạn")
			return
		}

		// 3. Trích xuất thông tin (Claims) và truyền vào Context cho Controller dùng
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("role", claims["role"])
		}

		c.Next() // Cho phép đi tiếp vào Handler
	}
}

// RequireRole phân quyền RBAC (Role-Based Access Control)
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusForbidden, "Không thể xác thực quyền")
			return
		}

		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Error(c, http.StatusForbidden, "Bạn không có quyền truy cập")
			return
		}

		c.Next()
	}
}
