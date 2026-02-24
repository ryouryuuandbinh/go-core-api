package media

import (
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"go-core-api/pkg/custom_error"

	"github.com/google/uuid"
)

// Giới hạn độ phân giải ảnh để chống Image Bomb (Pixel Flood)
const MaxImageWidth = 4096
const MaxImageHeight = 4096

func SaveAndProcessImage(file *multipart.FileHeader, dst string) (string, error) {
	openedFile, err := file.Open()
	if err != nil {
		return "", custom_error.ErrUploadFailed
	}
	defer openedFile.Close()

	// SECURITY FIX: Kiểm tra kích thước (Dimensions) trước khi cấp phát bộ nhớ RAM
	// DecodeConfig siêu nhẹ vì nó chỉ đọc Header của file
	config, format, err := image.DecodeConfig(openedFile)
	if err != nil {
		return "", custom_error.ErrCorruptedFile
	}

	if format != "jpeg" && format != "png" && format != "gif" {
		return "", custom_error.ErrInvalidFileType
	}

	if config.Width > MaxImageWidth || config.Height > MaxImageHeight {
		return "", custom_error.ErrFileTooLarge
	}

	openedFile.Seek(0, 0)
	img, _, err := image.Decode(openedFile)
	if err != nil {
		return "", custom_error.ErrCorruptedFile
	}

	subFolder := time.Now().Format("2006/01/02")
	uploadDir := filepath.Join(dst, subFolder)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", custom_error.ErrUploadFailed
	}

	newFileName := uuid.New().String() + ".jpg"
	finalPath := filepath.Join(uploadDir, newFileName)

	out, err := os.Create(finalPath)
	if err != nil {
		return "", custom_error.ErrUploadFailed
	}
	defer out.Close()

	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", custom_error.ErrUploadFailed
	}

	return filepath.ToSlash(finalPath), nil
}
