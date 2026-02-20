package services

import (
	"errors"
	"math"

	"go-core-api/internal/models"
	"go-core-api/internal/repositories"
	"go-core-api/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetListUsers(pagination utils.Pagination) ([]models.User, int64, int, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	GetProfile(userID uint) (*models.User, error)
	UpdateProfile(userID uint, fullName, avatar, phone string) error
	GetUserByID(id uint) (*models.User, error)
	AdminUpdateUser(id uint, role string) error
	DeleteUser(id uint) error
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

// GetListUsers xử lý logic tính toán tổng số trang

func (s *userService) GetListUsers(pagination utils.Pagination) ([]models.User, int64, int, error) {
	users, total, err := s.repo.GetList(pagination)
	if err != nil {
		return nil, 0, 0, err
	}

	// Tính tổng số trang = Ceil(Total / Limit)
	// Ví dụ: 15 records / 10 = 1.5 -> làm tròn lên 2 trang
	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return users, total, totalPages, nil
}

// ChangePassword xử lý logic kiểm tra và đổi mật khẩu
func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 1. Lấy thông tin từ DB
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("không tìm thấy người dùng")
	}

	// 2. Kiểm tra mật khẩu cũ xem có khớp không
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("mật khẩu cũ không chính xác")
	}

	// 3. Mã hoá mật khẩu mới
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("lỗi mã hoá mật khẩu mới")
	}

	// 4. Lưu vào database
	user.Password = string(hashedPassword)
	return s.repo.Update(user)
}

// GetProfile lấy thông tin chi tiết của 1 user
func (s *userService) GetProfile(userID uint) (*models.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("không tìm thấy người dùng")
	}

	return user, nil
}

// UpdateProfile cập nhật thông tin cá nhân của user
func (s *userService) UpdateProfile(userID uint, fullName, avatar, phone string) error {
	// 1. Tìm user hiện tại
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errors.New("không tìm thấy người dùng")
	}

	// 2. Cập nhật các trường dữ liệu
	user.FullName = fullName
	user.Avatar = avatar
	user.Phone = phone

	// 3. Lưu lại vào Database (Đã có sẵn hàm Update ở bài trước)
	return s.repo.Update(user)
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("không tìm thấy người dùng")
	}
	return user, nil
}

// Cập nhật thông tin User
func (s *userService) AdminUpdateUser(id uint, role string) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("không tìm thấy người dùng")
	}

	// Validate role
	if role != models.RoleAdmin && role != models.RoleUser {
		return errors.New("quyền không hợp lệ (chỉnhận 'admin' hoặc 'user')")
	}

	user.Role = role
	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("không tìm thấy người dùng để xoá")
	}

	return s.repo.Delete(id)
}
