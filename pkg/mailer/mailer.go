package mailer

import (
	"gopkg.in/gomail.v2"
)

// Mailer interface giúp dễ dàng mở rộng hoặc Mock test sau này
type Mailer interface {
	SendMail(to string, subject string, body string) error
}

type mailer struct {
	dialer *gomail.Dialer
	from   string
}

// NewMailer khởi tạo kết nối SMTP
func NewMailer(host string, port int, user string, password string, from string) Mailer {
	dialer := gomail.NewDialer(host, port, user, password)
	return &mailer{
		dialer: dialer,
		from:   from,
	}
}

func (m *mailer) SendMail(to string, subject string, body string) error {
	msg := gomail.NewMessage()

	// Thiết lập Header
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)

	// Nội dung email (text/html cho phép gửi mail đẹp có màu sắc, hình ảnh)
	msg.SetBody("text/html", body)

	// --- THÊM ĐOẠN NÀY ĐỂ DEBUG ---
	// In ra xem code nó đang gửi cái gì đi
	// (Nhớ import "fmt" ở đầu file)
	// fmt.Printf("📧 Đang gửi mail:\n - From: %s\n - To: %s\n - Host: %s\n", m.from, to, m.dialer.Host)
	// ------------------------------

	// Thực hiện gửi
	if err := m.dialer.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}
