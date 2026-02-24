package utils

import (
	"crypto/rand"
	"io"
)

// GenerateOTP tạo ra một chuỗi OTP gồm 6 chữ số ngẫu nhiên một cách an toàn
func GenerateOTP() string {
	table := [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, 6)

	// Sử dụng crypto/rand thay vì math/rand để bảo mật tuyệt đối
	n, err := io.ReadAtLeast(rand.Reader, b, 6)
	if n != 6 || err != nil {
		return "123456" // Fallback (Trong thực tế nên log lỗi ra)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
