package media

import (
	"errors"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Giới hạn độ phân giải ảnh để chống Image Bomb (Pixel Flood)
const MaxImageWidth = 4096
const MaxImageHeight = 4096

func SaveAndProcessImage(file *multipart.FileHeader, dst string) (string, error) {
	openedFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer openedFile.Close()

	// SECURITY FIX: Kiểm tra kích thước (Dimensions) trước khi cấp phát bộ nhớ RAM
	// DecodeConfig siêu nhẹ vì nó chỉ đọc Header của file
	config, format, err := image.DecodeConfig(openedFile)
	if err != nil {
		return "", errors.New("file tải lên không hợp lệ hoặc bị hỏng")
	}

	if format != "jpeg" && format != "png" && format != "gif" {
		return "", errors.New("chỉ hỗ trợ định dạng JPEG, PNG và GIF")
	}

	if config.Width > MaxImageWidth || config.Height > MaxImageHeight {
		return "", errors.New("kích thước ảnh quá lớn, tối đa 4096x4096px")
	}

	// Reset lại con trỏ file về đầu (0) để bắt đầu Decode thực sự
	openedFile.Seek(0, 0)

	// Decode thực sự đưa vào RAM (Lúc này đã an toàn)
	img, _, err := image.Decode(openedFile)
	if err != nil {
		return "", errors.New("lỗi trong quá trình đọc ảnh")
	}

	// Chuẩn bị thư mục
	subFolder := time.Now().Format("2006/01/02")
	uploadDir := filepath.Join(dst, subFolder)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}

	newFileName := uuid.New().String() + ".jpg"
	finalPath := filepath.Join(uploadDir, newFileName)

	out, err := os.Create(finalPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Encode lại ảnh sang JPEG để khử mã độc
	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", errors.New("lỗi trong quá trình xử lý ảnh")
	}

	return filepath.ToSlash(finalPath), nil
}
