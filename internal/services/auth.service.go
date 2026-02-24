// Package services xử lý các logic nghiệp vụ của hệ thống
package services

import (
	"context"
	"errors"
	"time"

	"go-core-api/internal/models"
	"go-core-api/internal/repositories"
	"go-core-api/pkg/config"
	"go-core-api/pkg/logger"
	"go-core-api/pkg/mailer"
	"go-core-api/pkg/utils"
	templates "go-core-api/template"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// TokenDetails chứa thông tin về AccessToken và RefreshToken sau khi login
type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService interface {
	Register(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (*TokenDetails, error)
	GenerateTokens(userID uint, role string, tokenVersion int) (*TokenDetails, error)
	RefreshToken(ctx context.Context, tokenString string) (*TokenDetails, error)
	RevokeToken(ctx context.Context, userID uint) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
}

type authService struct {
	repo   repositories.UserRepository
	secret string
	mailer mailer.Mailer
}

func NewAuthService(repo repositories.UserRepository, secret string, mail mailer.Mailer) AuthService {
	return &authService{
		repo:   repo,
		secret: secret,
		mailer: mail,
	}
}

// THUẬT TOÁN ĐĂNG KÝ: Hash password bằng bcrypt với độ khó (cost) = 10
func (s *authService) Register(ctx context.Context, email, password string) error {
	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return errors.New("email đã được sử dụng")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     models.RoleUser,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}

	// Kích hoạt Event gửi mail ngay trong Service (Clean Code)
	utils.RunInBackground(func() {
		subject := "🎉 Welcome to [YourApp]!"

		// Bơm dữ liệu vào template welcome.html
		body, err := templates.Render("welcome.html", map[string]interface{}{
			"Email": email,
			"Link":  config.AppConfig.Server.Domain,
		})

		if err != nil {
			logger.Error("Lỗi render template welcome", zap.Error(err))
			return
		}

		if err := s.mailer.SendMail(email, subject, body); err != nil {
			logger.Error("Lỗi gửi email chào mừng", zap.Error(err))
		}
	})

	return nil
}

// THUẬT TOÁN LOGIN & JWT
func (s *authService) Login(ctx context.Context, email, password string) (*TokenDetails, error) {
	// 1. Tìm user
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("sai email hoặc mật khẩu")
	}

	// 2. So sánh mật khẩu người dùng nhập với mật khẩu hash trong DB
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("sai email hoặc mật khẩu")
	}

	// 3. Cấp phát Token
	return s.GenerateTokens(user.ID, user.Role, user.TokenVersion)
}

// Logic sinh cặp Token (Access & RefreshToken)
func (s *authService) GenerateTokens(userID uint, role string, tokenVersion int) (*TokenDetails, error) {
	cfg := config.AppConfig.JWT

	// Access Token dùng cấu hình AccessExpiration
	accessTokenClaims := jwt.MapClaims{
		"token_type":    "access",
		"user_id":       userID,
		"role":          role,
		"token_version": tokenVersion,
		"exp":           time.Now().Add(time.Minute * time.Duration(cfg.AccessExpiration)).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	aToken, err := accessToken.SignedString([]byte(s.secret))
	if err != nil {
		return nil, err
	}

	// Refresh Token dùng cấu hình RefreshExpiration
	refreshTokenClaim := jwt.MapClaims{
		"token_type": "refresh",
		"user_id":    userID,
		"exp":        time.Now().Add(time.Hour * 24 * time.Duration(cfg.RefreshExpiration)).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaim)
	rToken, err := refreshToken.SignedString([]byte(s.secret))
	if err != nil {
		return nil, err
	}
	return &TokenDetails{
		AccessToken:  aToken,
		RefreshToken: rToken,
	}, nil
}

// RefreshToken giải mã token cũ và cấp phát token mới
func (s *authService) RefreshToken(ctx context.Context, tokenString string) (*TokenDetails, error) {
	// 1. Giải mã và kiểm tra tính hợp lệ của Refresh Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("refresh token không hợp lệ hoặc đã hết hạn")
	}

	// 2. Trích xuất thông tin user_id từ token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("không thể đọc thông tin token")
	}

	// Kiểm tra xem có phải là refresh token không
	if claims["token_type"] != "refresh" {
		return nil, errors.New("token không phải là refresh token")
	}

	// Lưu ý: jwt lưu số dưới dạng float64, nên phải ép kiểu cẩn thận
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("token sai định dạng")
	}

	userID := uint(userIDFloat)

	// 3. Kiểm tra xem User này còn tồn tại trong DB không
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("tài khoản không tồn tại")
	}

	// 4. Nếu mọi thứ OK, tạo cặp Token mới dựa vào ID và Role của User
	return s.GenerateTokens(user.ID, user.Role, user.TokenVersion)
}

func (s *authService) RevokeToken(ctx context.Context, userID uint) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("không tìm thấy người dùng")
	}
	// Tăng TokenVersion khiến mọi JWT hiện tại trở thành vô nghĩa
	user.TokenVersion += 1
	return s.repo.Update(ctx, user)
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		// [REFACTOR - BẢO MẬT] Trả về "nil" thay vì lỗi để tránh bị hacker "dò" (Enumeration) xem email nào có trong hệ thống
		return nil
	}

	// Tạo Token ngẫu nhiên bằng UUID
	otpCode := utils.GenerateOTP()
	expiry := time.Now().Add(15 * time.Minute)

	user.ResetPasswordOTP = &otpCode
	user.ResetPasswordExpires = &expiry

	if err := s.repo.Update(ctx, user); err != nil {
		return errors.New("lỗi hệ thống khi tạo mã khôi phục")
	}

	utils.RunInBackground(func() {
		subject := "🔑 Your Password Reset Code"

		// Bơm mã OTP vào template reset_password.html
		body, err := templates.Render("reset_password.html", map[string]interface{}{
			"OTP": otpCode,
		})

		if err != nil {
			logger.Error("Lỗi render template reset password", zap.Error(err))
			return
		}

		if err := s.mailer.SendMail(user.Email, subject, body); err != nil {
			logger.Error("Lỗi gửi email khôi phục", zap.Error(err))
		}
	})

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, OTP string, newPassword string) error {
	user, err := s.repo.FindByResetOTP(ctx, OTP)
	if err != nil {
		return errors.New("mã OTP không hợp lệ")
	}

	// Kiểm tra hết hạn
	if user.ResetPasswordExpires == nil || user.ResetPasswordExpires.Before(time.Now()) {
		return errors.New("mã OTP đã hết hạn")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("lỗi mã hoá mật khẩu mới")
	}

	user.Password = string(hashedPassword)
	user.ResetPasswordOTP = nil // Xóa token sau khi dùng
	user.ResetPasswordExpires = nil
	user.TokenVersion += 1

	return s.repo.Update(ctx, user)
}
