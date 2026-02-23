package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"go-core-api/internal/repositories"
	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(secret string, userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Yêu cầu đăng nhập")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, http.StatusUnauthorized, "Token sai định dạng")
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("phương thức mã hoá không hợp lệ")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, http.StatusUnauthorized, "Token hết hạn hoặc bị can thiệp")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["token_type"] != "access" {
			response.Error(c, http.StatusUnauthorized, "Sử dụng sai loại Token")
			c.Abort()
			return
		}

		// BUG FIX: Ép kiểu an toàn, tránh Panic làm sập server
		userIDFloat, okID := claims["user_id"].(float64)
		tokenVersionFloat, okVer := claims["token_version"].(float64)

		if !okID || !okVer {
			response.Error(c, http.StatusUnauthorized, "Payload của Token không hợp lệ")
			c.Abort()
			return
		}

		userID := uint(userIDFloat)
		tokenVersion := int(tokenVersionFloat)

		user, err := userRepo.FindByID(c.Request.Context(), userID)
		if err != nil || user.TokenVersion != tokenVersion {
			response.Error(c, http.StatusUnauthorized, "Phiên đăng nhập đã hết hạn (vui lòng đăng nhập lại)")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", claims["role"])
		c.Next()
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
