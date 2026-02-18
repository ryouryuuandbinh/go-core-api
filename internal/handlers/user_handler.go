package handlers

import (
	"net/http"

	"go-core-api/internal/services"
	"go-core-api/pkg/response"
	"go-core-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetList lấy danh sách user có phân trang

func (h *UserHandler) GetList(c *gin.Context) {
	// 1. Lấy tham số phân trang từ UserHandler
	pagination := utils.GeneratePaginationFromRequest(c)

	// 2. Gọi Service
	users, total, totalPages, err := h.service.GetListUsers(pagination)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Lỗi lấy danh sách")
		return
	}

	// 3. Trả về response chuẩn có Meta Data
	response.Success(c, http.StatusOK, "Lấy danh sách thành công", gin.H{
		"items": users,
		"meta": gin.H{
			"total":       total,
			"total_pages": totalPages,
			"page":        pagination.Page,
			"limit":       pagination.Limit,
			"sort":        pagination.Sort,
			"keyword":     pagination.Keyword,
		},
	})
}
