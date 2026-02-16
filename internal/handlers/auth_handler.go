package handlers

import (
	"net/http"

	"go-core-api/internal/services"
	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service services.AuthService
	secret  string // Load từ config truyền vào
}

func NewAuthHandler(service services.AuthService, secret string) *AuthHandler {
	return &AuthHandler{service, secret}
}

type AuthRequest struct {
	Email    string `json:"email" binding:"required, email"`
	Password string `json:"password" binding:"required, min=6`
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
