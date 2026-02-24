package custom_error

import "net/http"

// AppError định nghĩa cấu trúc lỗi chuẩn của toàn hệ thống
type AppError struct {
	HTTPCode int    `json:"-"`        // Không trả ra JSON, chỉ dùng để set HTTP Status Code
	Code     string `json:"err_code"` // Mã lỗi cho Frontend (VD: ERR_USER_404)
	Message  string `json:"message"`  // Thông báo lỗi cho người dùng đọc
}

// Hàm Error() giúp AppError thoả mãn interface error mặc định của Golang
func (e *AppError) Error() string {
	return e.Message
}

// New khởi tạo một lỗi mới
func New(httpCode int, code, message string) *AppError {
	return &AppError{
		HTTPCode: httpCode,
		Code:     code,
		Message:  message,
	}
}

// ==============================================================================
// DANH SÁCH CÁC MÃ LỖI ĐƯỢC ĐỊNH NGHĨA SẴN (PRE-DEFINED ERRORS)
// ==============================================================================

var (
	// Lỗi hệ thống & Validate
	ErrInvalidRequest  = New(http.StatusBadRequest, "ERR_BAD_REQUEST", "Dữ liệu yêu cầu không hợp lệ")
	ErrUnauthorized    = New(http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Không có quyền truy cập hoặc phiên đăng nhập hết hạn")
	ErrForbidden       = New(http.StatusForbidden, "ERR_FORBIDDEN", "Bạn không có quyền thực hiện hành động này")
	ErrInternalServer  = New(http.StatusInternalServerError, "ERR_INTERNAL_SERVER", "Lỗi hệ thống, vui lòng thử lại sau")
	ErrTooManyRequests = New(http.StatusTooManyRequests, "ERR_TOO_MANY_REQUESTS", "Bạn đã gửi quá nhiều yêu cầu. Vui lòng thử lại sau")

	// Lỗi liên quan đến User & Auth
	ErrUserNotFound       = New(http.StatusNotFound, "ERR_USER_NOT_FOUND", "Không tìm thấy người dùng")
	ErrEmailExists        = New(http.StatusConflict, "ERR_EMAIL_EXISTS", "Email đã được sử dụng")
	ErrInvalidCredentials = New(http.StatusUnauthorized, "ERR_INVALID_CREDENTIALS", "Sai email hoặc mật khẩu")
	ErrWrongPassword      = New(http.StatusBadRequest, "ERR_WRONG_PASSWORD", "Mật khẩu cũ không chính xác")
	ErrInvalidOTP         = New(http.StatusBadRequest, "ERR_INVALID_OTP", "Mã OTP không chính xác")
	ErrOTPExpired         = New(http.StatusBadRequest, "ERR_OTP_EXPIRED", "Mã OTP đã hết hạn")
	ErrCannotDeleteSelf   = New(http.StatusForbidden, "ERR_CANNOT_DELETE_SELF", "Hành động nguy hiểm: Không thể tự xoá chính mình")

	// Lỗi Media & Upload
	ErrUploadFailed    = New(http.StatusInternalServerError, "ERR_UPLOAD_FAILED", "Lỗi trong quá trình xử lý file")
	ErrFileTooLarge    = New(http.StatusRequestEntityTooLarge, "ERR_FILE_TOO_LARGE", "Dung lượng file vượt quá giới hạn (Tối đa 5MB)")
	ErrInvalidFileType = New(http.StatusBadRequest, "ERR_INVALID_FILE_TYPE", "Chỉ hỗ trợ định dạng JPEG, PNG và GIF")
	ErrCorruptedFile   = New(http.StatusBadRequest, "ERR_CORRUPTED_FILE", "File tải lên không hợp lệ hoặc bị hỏng")
)
