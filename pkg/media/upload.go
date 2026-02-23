package media

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// c: Context của Gin
// file: Header của file gửi lên
// dst: Đường đãn thư mục muốn lưu (ví dụ: "uploads/avatars")
// SaveFile xử lý lưu file từ request gửi lên

func SaveFile(c *gin.Context, file *multipart.FileHeader, dst string) (string, error) {
	// 1. Mở file để kiểm tra nội dung thực sự (Bảo mật MIME Type)
	openedFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer openedFile.Close()
	// Đọc 512 byte đầu tiên để Gof nhận diện định dạng
	buffer := make([]byte, 512)
	if _, err := openedFile.Read(buffer); err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)

	// Các MIME type ảnh hợp lệ
	allowTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true, // Thêm webp cho hiện đại
	}
	if !allowTypes[contentType] {
		return "", fmt.Errorf("định dạng file thực sự không hỗ trợ (phát hiện: %s)", contentType)
	}

	// 2. Lấy extension để lưu tên file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg" // Fallback nếu file không có đuôi
	}

	// 2. Tạo cấu trúc thư mục theo ngày: YYYY/MM/DD
	subFolder := time.Now().Format("2006/01/02")

	// Đường dẫn lưu thực tế: uploads/2026/02/18
	uploadDir := filepath.Join(dst, subFolder)

	// 3. Tạo thư mục nếu chưa tồn tại
	// os.MkdirAll sẽ tạo cả thư mục cha nếu cần (VD: uploads/2026/02)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}

	// 4. Tạo tên file mới ngẫu nhiên (UUID) để tránh trùng
	newFileName := uuid.New().String() + ext

	// Đường dẫn lưu file cuối cùng: uploads/2026/02/18/uuid-cua-ban.jpg
	finalPath := filepath.Join(uploadDir, newFileName)

	// 5. Lưu file (dùng hàm có sẵn của Gin)
	if err := c.SaveUploadedFile(file, finalPath); err != nil {
		return "", err
	}

	// Trả về đường dẫn để lưu vào DB (Lưu ý: đổi dấu \ thành / nếu chạy trên Windows đẻ URL hợp lệ)
	return filepath.ToSlash(finalPath), nil
}
