package templates

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed *.html
var emailTmplFS embed.FS

var tmpl *template.Template

// Hàm init() tự động chạy một lần duy nhất khi Server khởi động
func init() {
	// Parse toàn bộ các file .html có trong thư mục này
	tmpl = template.Must(template.ParseFS(emailTmplFS, "*.html"))
}

// Render là hàm dùng để fill dữ liệu (data) vào một file HTML cụ thể
func Render(filename string, data interface{}) (string, error) {
	var buf bytes.Buffer
	// Trộn data vào file template và ghi vào buffer
	if err := tmpl.ExecuteTemplate(&buf, filename, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
