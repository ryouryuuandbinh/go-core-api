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
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" { // Case-insensitive cho Bearer
			response.Error(c, http.StatusUnauthorized, "Token sai định dạng")
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
			response.Error(c, http.StatusUnauthorized, "Token hết hạn hoặc không hợp lệ")
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if claims["token_type"] != "access" {
				response.Error(c, http.StatusUnauthorized, "Sử dụng sai loại Token")
				return
			}

			// JWT lưu số (number) dưới dạng float64, cần ép kiểu an toàn
			userIDFloat, _ := claims["user_id"].(float64)
			userID := uint(userIDFloat)

			tokenVersionFloat, _ := claims["token_version"].(float64)
			tokenVersion := int(tokenVersionFloat)

			// --- KIỂM TRA BẢO MẬT: So sánh Token Version với Database ---
			// (Lưu ý: Để tối ưu hiệu năng cho hệ thống lớn, đoạn này nên query từ Redis Cache thay vì Postgres)
			user, err := userRepo.FindByID(c.Request.Context(), userID)
			if err != nil || user.TokenVersion != tokenVersion {
				response.Error(c, http.StatusUnauthorized, "Phiên đăng nhập đã hết hạn hoặc bị thu hồi (vui lòng đăng nhập lại)")
				return
			}

			c.Set("user_id", userID)
			c.Set("role", claims["role"])
		}

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
