// Package services xử lý các logic nghiệp vụ của hệ thống
package services

import (
	"errors"
	"time"

	"go-core-api/internal/models"
	"go-core-api/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TokenDetails chứa thông tin về AccessToken và RefreshToken sau khi login
type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	Register(email, password string) error
	Login(email, password, secret string) (*TokenDetails, error)
	GenerateTokens(userID uint, role, secret string) (*TokenDetails, error)
}

type authService struct {
	repo repositories.UserRepository
}

func NewAuthService(repo repositories.UserRepository) AuthService {
	return &authService{repo: repo}
}

// THUẬT TOÁN ĐĂNG KÝ: Hash password bằng bcrypt với độ khó (cost) = 10
func (s *authService) Register(email, password string) error {
	// 1. Kiểm tra Email đã tồn tại chưa
	// Nếu err == nil nghĩa là tìm thấy user -> Trùng email -> Báo lỗi
	if _, err := s.repo.FindByEmail(email); err == nil {
		return errors.New("email đã được sử dụng")
	}

	// 2. Mã hoá mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 3. Lưu vào DB
	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     models.RoleUser,
	}
	return s.repo.Create(user)
}

// THUẬT TOÁN LOGIN & JWT
func (s *authService) Login(email, password, secret string) (*TokenDetails, error) {
	// 1. Tìm user
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("sai email hoặc mật khẩu")
	}

	// 2. So sánh mật khẩu người dùng nhập với mật khẩu hash trong DB
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("sai email hoặc mật khẩu")
	}

	// 3. Cấp phát Token
	return s.GenerateTokens(user.ID, user.Role, secret)
}

// Logic sinh cặp Token (Access & RefreshToken)
func (s *authService) GenerateTokens(userID uint, role, secret string) (*TokenDetails, error) {
	// Access Token tuổi thọ ngắn (15 phút), dùng để call API liên tục
	accessTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	aToken, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	// Refresh Token tuổi thọ dài (7 ngày), dùng để xin lại Access Token mới khi cái cũ hết hạn
	refreshTokenClaim := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaim)
	rToken, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}
	return &TokenDetails{
		AccessToken:  aToken,
		RefreshToken: rToken,
	}, nil
}
