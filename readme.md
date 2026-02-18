# Go Core API Boilerplate (Clean Architecture)

Đây là bộ khung Backend API chuẩn Enterprise, được xây dựng với mục đích học tập và làm nền tảng cho các dự án lớn.

## 🛠 Tech Stack
- **Language:** Go (Golang)
- **Framework:** Gin Gonic
- **Database:** PostgreSQL
- **ORM:** GORM
- **Authentication:** JWT (Access & Refresh Token)
- **Architecture:** Clean Architecture (Handler -> Service -> Repository)

## 🚀 Chức năng hiện tại
1. **Kiến trúc chuẩn:** Phân chia thư mục rõ ràng, dễ mở rộng.
2. **Database:** Kết nối Postgres với Connection Pool tối ưu.
3. **Authentication:**
   - Đăng ký / Đăng nhập (Hash password với Bcrypt).
   - Middleware xác thực JWT.
   - Cơ chế bảo vệ Route theo Role (RBAC).

## ⚙️ Cài đặt & Chạy
1. Clone dự án.
2. Copy `config/config.example.yaml` thành `config/config.yaml`.
3. Cập nhật thông tin Database trong file config.
4. Chạy lệnh:
   ```bash
   go mod tidy
   go run cmd/main.go
