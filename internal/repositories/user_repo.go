package repositories

import (
	"go-core-api/internal/models"
	"go-core-api/pkg/utils"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	GetList(pagination utils.Pagination) ([]models.User, int64, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetList(pagination utils.Pagination) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// 1. Tìm kiếm (Filtering)
	// Nếu có keyword, tìm theo Email (hoặc tên)
	if pagination.Keyword != "" {
		// Dấu % là cú pháp SQL: tìm chuỗi chứa Keyword (LIKE %abc%)
		query = query.Where("email ILIKE ?", "%"+pagination.Keyword+"%")
	}

	// 2. Đếm tổng số bản ghi (quan trọng để frontend vẽ nút phân trang)
	// Models(&models.User{}) báo cho GORM biết đang làm việc với bảng users
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 3. Lấy dữ liệu phân trang
	// Offset: Bỏ qua bao nhiêu dòng
	// Limit: lấy bao nhiêu dòng
	offset := (pagination.Page - 1) * pagination.Limit

	err := query.Limit(pagination.Limit).
		Offset(offset).
		Order(pagination.Sort).
		Find(&users).Error

	return users, total, err
}

func (r *userRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepo) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *userRepo) Update(user *models.User) error {
	// Dùng save để cập nhật toàn bộ thông tin của user hiện tại
	return r.db.Save(user).Error
}

func (r *userRepo) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}
