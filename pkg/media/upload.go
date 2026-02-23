package media

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// c: Context của Gin
// file: Header của file gửi lên
// dst: Đường đãn thư mục muốn lưu (ví dụ: "uploads/avatars")
// SaveFile xử lý lưu file từ request gửi lên

func SaveFile(c *gin.Context, file *multipart.FileHeader, dst string) (string, error) {
	openedFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer openedFile.Close()

	buffer := make([]byte, 512)
	n, err := openedFile.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	if n == 0 {
		return "", fmt.Errorf("không thể tải lên file rỗng")
	}
	contentType := http.DetectContentType(buffer[:n])

	extMap := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}

	ext, isValid := extMap[contentType]
	if !isValid {
		return "", fmt.Errorf("định dạng file không hỗ trợ: %s", contentType)
	}

	subFolder := time.Now().Format("2006/01/02")
	uploadDir := filepath.Join(dst, subFolder)

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}

	newFileName := uuid.New().String() + ext
	finalPath := filepath.Join(uploadDir, newFileName)

	if err := c.SaveUploadedFile(file, finalPath); err != nil {
		return "", err
	}

	return filepath.ToSlash(finalPath), nil
}
