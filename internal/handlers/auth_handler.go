package handlers

import (
	"net/http"

	"go-core-api/internal/services"
	"go-core-api/pkg/mailer"
	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service services.AuthService
	mailer  mailer.Mailer
	secret  string // Load từ config truyền vào
}

func NewAuthHandler(service services.AuthService, mailer mailer.Mailer, secret string) *AuthHandler {
	return &AuthHandler{
		service: service,
		mailer:  mailer,
		secret:  secret,
	}
}

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req AuthRequest
	// Validation dữ liệu đầu vào
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "dữ liệu không hợp lệ")
		return
	}

	err := h.service.Register(req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	// 2. Gửi Email chào mừng (BẤT ĐỒNG BỘ - ASYNC)
	go func() {
		// Import thêm package "fmt" ở đầu file nếu chưa có
		subject := "Chào mừng thành viên mới!"
		body := "<h1>Xin chào " + req.Email + "</h1><p>Cảm ơn bạn đã tham gia.</p>"

		_ = h.mailer.SendMail(req.Email, subject, body)
		// err := h.mailer.SendMail(req.Email, subject, body)
		// if err != nil {
		// 	// In lỗi đỏ lòm ra màn hình cho dễ thấy
		// 	fmt.Printf("❌ LỖI GỬI MAIL: %v\n", err)
		// } else {
		// 	fmt.Println("✅ Đã gửi mail thành công!")
		// }
	}()
	response.Success(c, http.StatusCreated, "Đăng ký thành công", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	tokens, err := h.service.Login(req.Email, req.Password, h.secret)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Đăng nhập thành công", tokens)
}

// RefreshToken xử lý requét cấp lại token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Vui lòng cung cấp refresh_token")
		return
	}

	// Gọi Service
	tokens, err := h.service.RefreshToken(req.RefreshToken, h.secret)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Làm mới token thành công", tokens)
}
