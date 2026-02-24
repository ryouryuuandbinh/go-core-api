package handlers

import (
	"net/http"

	"go-core-api/internal/services"
	"go-core-api/pkg/response"
	"go-core-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service services.AuthService
}

func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
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

func (h *AuthHandler) Logout(c *gin.Context) {
	// Trích xuất userID từ Access Token hiện tại
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Không thể xác định người dùng")
		return
	}

	// Gọi Service hủy Token
	if err := h.service.RevokeToken(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi hệ thống khi đăng xuất")
		return
	}

	response.Success(c, http.StatusOK, "Đăng xuất an toàn trên toàn hệ thống", nil)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	// Gọi service
	err := h.service.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Luôn trả về thành công dù email có tồn tại hay không
	response.Success(c, http.StatusOK, "Nếu email hợp lệ, hướng dẫn khôi phục đã được gửi tới hòm thư của bạn.", nil)
}

// [BỔ SUNG] API Submit Reset Password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ. Mật khẩu tối thiểu 8 ký tự.")
		return
	}

	err := h.service.ResetPassword(c.Request.Context(), req.OTP, req.NewPassword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Đặt lại mật khẩu thành công. Vui lòng đăng nhập lại.", nil)
}
