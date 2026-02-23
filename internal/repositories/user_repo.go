package repositories

import (
	"context"
	"go-core-api/internal/models"
	"go-core-api/pkg/utils"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	GetList(ctx context.Context, pagination utils.Pagination) ([]models.User, int64, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return &user, err
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepo) GetList(ctx context.Context, pagination utils.Pagination) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

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

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	// Dùng save để cập nhật toàn bộ thông tin của user hiện tại
	return r.db.WithContext(ctx).Updates(user).Error
}

func (r *userRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}
