package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/higgstv/higgstv-go/internal/models"
	"github.com/higgstv/higgstv-go/internal/repository"
	"github.com/higgstv/higgstv-go/pkg/uuidutil"
)

// AuthService 認證服務
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService 建立認證服務
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// SignIn 登入
func (s *AuthService) SignIn(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// SignUp 註冊
func (s *AuthService) SignUp(ctx context.Context, invitationCode, username, email, password string) (*models.User, error) {
	// 驗證邀請碼
	if invitationCode != "sixpens" {
		return nil, errors.New("invalid invitation code")
	}

	// 檢查使用者是否存在
	exists, err := s.userRepo.Exists(ctx, username, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user already exists")
	}

	// 加密密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 建立使用者
	user := &models.User{
		ID:       uuidutil.NewBase64UUID(),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword 變更密碼
func (s *AuthService) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 驗證舊密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// 加密新密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密碼
	return s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
}

// GenerateAccessKey 產生 access_key（用於重設密碼）
func (s *AuthService) GenerateAccessKey(ctx context.Context, email string) (string, error) {
	// 產生隨機 access_key
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	accessKey := base64.URLEncoding.EncodeToString(bytes)

	// 儲存到資料庫
	if err := s.userRepo.SetAccessKey(ctx, email, accessKey); err != nil {
		return "", err
	}

	return accessKey, nil
}

// ResetPassword 重設密碼
func (s *AuthService) ResetPassword(ctx context.Context, email, accessKey, newPassword string) error {
	// 加密新密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 使用 access_key 更新密碼
	success, err := s.userRepo.ChangePasswordWithAccessKey(ctx, email, accessKey, string(hashedPassword))
	if err != nil {
		return err
	}
	if !success {
		return errors.New("invalid access key")
	}

	return nil
}

