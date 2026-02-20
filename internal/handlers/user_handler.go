package handlers

import (
	"net/http"
	"strconv"

	"go-core-api/internal/services"
	"go-core-api/pkg/response"
	"go-core-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserHandler struct {
	service services.UserService
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone"`
}

type AdminUpdateUserRequest struct {
	Role string `json:"role" binding:"required"`
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

// ChangePassword xử lý request đổi mật khẩu của user đang đăng nhập
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ, mật khẩu mới phải từ 6 ký tự")
		return
	}

	// lấy user_id từ Context (do Middleware RequireAuth truyền vào)
	// Lưu ý: Middleware đang set user_id là kiểu float64 (chuẩn của JWT) nên cần ép kiểu
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	userID := uint(userIDFloat.(float64))

	// Gọi Service xử lý
	err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Đổi mật khẩu thành công", nil)
}

// UpdateProfile xử lý yêu cầu cập nhật hồ sơ
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	// 1. Lấy user_id từ Token
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Không tìm thấy thông tin xác thực")
		return
	}
	userID := uint(userIDFloat.(float64))

	// 2. Gọi Service để cập nhật
	err := h.service.UpdateProfile(userID, req.FullName, req.Avatar, req.Phone)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Cập nhật hồ sơ thành công", nil)
}

// GET /api/v1/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	// Lấy tham số :id URL (ví dụ /users/5 -> id = 5)
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	user, err := h.service.GetUserByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Thành công", user)
}

// PUT /api/v1/users/:id
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}

	if err := h.service.AdminUpdateUser(uint(id), req.Role); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Cập nhật quyền thành công", nil)
}

// DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	if err := h.service.DeleteUser(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Xóa người dùng thành công", nil)
}
