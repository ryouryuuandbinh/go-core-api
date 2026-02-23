package handlers

import (
	"net/http"

	"go-core-api/internal/services"
	"go-core-api/pkg/logger"
	"go-core-api/pkg/mailer"
	"go-core-api/pkg/response"
	"go-core-api/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

	err := h.service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}

	// 2. Gửi Email chào mừng (BẤT ĐỒNG BỘ)
	utils.RunInBackground(func() {
		subject := "Chào mừng thành viên mới!"
		body := "<h1>Xin chào " + req.Email + "</h1><p>Cảm ơn bạn đã tham gia.</p>"

		err := h.mailer.SendMail(req.Email, subject, body)
		if err != nil {
			// KHÔNG ĐƯỢC dùng _ để bỏ qua lỗi, hãy dùng logger hệ thống để ghi nhận
			logger.Error("Lỗi gửi email chào mừng", zap.String("email", req.Email), zap.Error(err))
		}
	})

	response.Success(c, http.StatusCreated, "Đăng ký thành công", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
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
	tokens, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Làm mới token thành công", tokens)
}
