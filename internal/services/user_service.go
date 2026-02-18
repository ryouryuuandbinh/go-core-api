package services

import (
	"math"

	"go-core-api/internal/models"
	"go-core-api/internal/repositories"
	"go-core-api/pkg/utils"
)

type UserService interface {
	GetListUsers(pagination utils.Pagination) ([]models.User, int64, int, error)
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
