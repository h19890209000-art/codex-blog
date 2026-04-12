package service

import (
	"errors"
	"strings"
	"time"

	"ai-blog/backend/internal/config"
	"ai-blog/backend/internal/model"
	"ai-blog/backend/internal/repository"
	"ai-blog/backend/internal/support"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 负责管理员登录和密码管理。
type AuthService struct {
	cfg      config.AuthConfig
	userRepo repository.UserRepository
}

// NewAuthService 创建登录服务。
func NewAuthService(cfg config.AuthConfig, userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
	}
}

// Login 校验用户名密码并返回 token。
func (service *AuthService) Login(username string, password string) (map[string]any, error) {
	user, err := service.userRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码不正确")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码不正确")
	}

	token, err := support.GenerateToken(service.cfg.TokenSecret, support.AuthClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		ExpireAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"token": token,
		"user":  sanitizeUser(user),
	}, nil
}

// Profile 返回当前登录用户信息。
func (service *AuthService) Profile(userID int64) (map[string]any, error) {
	user, err := service.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"user": sanitizeUser(user),
	}, nil
}

// ChangePassword 修改管理员密码。
func (service *AuthService) ChangePassword(userID int64, oldPassword string, newPassword string) error {
	if len(strings.TrimSpace(newPassword)) < 6 {
		return errors.New("新密码至少需要 6 位")
	}

	user, err := service.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("旧密码不正确")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return service.userRepo.UpdatePassword(userID, string(hashedPassword))
}

func sanitizeUser(user model.User) map[string]any {
	return map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"avatar":   user.Avatar,
		"role":     user.Role,
	}
}
