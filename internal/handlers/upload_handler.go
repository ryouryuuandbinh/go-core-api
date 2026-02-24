package handlers

import (
	"net/http"

	"go-core-api/pkg/config"
	"go-core-api/pkg/custom_error"
	"go-core-api/pkg/media"
	"go-core-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

const MaxFileSize = 5 << 20

// UploadImage xử lý upload 1 ảnh
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// 1. Lấy file từ form-data (key là "file")
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, custom_error.ErrInvalidRequest)
		return
	}

	if file.Size > MaxFileSize {
		response.Error(c, custom_error.ErrFileTooLarge)
		return
	}

	// 2. Gọi hàm SaveFile trong pkg
	// Lưu vào thư mục "uploads"
	filePath, err := media.SaveAndProcessImage(file, "uploads")
	if err != nil {
		response.Error(c, err)
		return
	}

	// 3. Trả về đường dẫn file (cộng thêm domain server nếu cần)
	// Ví dụ: http://localhost:8080/uploads/uuid.jpg
	domain := config.AppConfig.Server.Domain
	fullURL := domain + "/" + filePath
	response.Success(c, http.StatusOK, "Upload thành công", gin.H{
		"url": fullURL,
	})
}
