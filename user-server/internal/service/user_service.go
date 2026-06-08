package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// UserService 用户管理服务.
type UserService struct {
	userRepo       UserRepository
	sessionRepo    SessionRepository
	verifyCodeRepo VerifyCodeRepository
}

// NewUserService 创建用户管理服务.
func NewUserService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	verifyCodeRepo VerifyCodeRepository,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		verifyCodeRepo: verifyCodeRepo,
	}
}

// GetUser 获取用户信息.
func (s *UserService) GetUser(ctx context.Context, userID string) (*schema.User, error) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		slog.Error("获取用户信息失败", "user_id", userID, "error", err)
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user, nil
}

// UpdateUserInfo 更新用户信息.
func (s *UserService) UpdateUserInfo(ctx context.Context, userID, name, email, avatar string) (*schema.User, error) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}
	if avatar != "" {
		user.Avatar = avatar
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		slog.Error("更新用户信息失败", "user_id", userID, "error", err)
		return nil, fmt.Errorf("更新用户信息失败: %w", err)
	}

	slog.Info("更新用户信息成功", "user_id", userID)
	return user, nil
}

// SetAvatar 设置用户头像.
func (s *UserService) SetAvatar(ctx context.Context, userID, avatar string) error {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Avatar = avatar
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		slog.Error("设置头像失败", "user_id", userID, "error", err)
		return fmt.Errorf("设置头像失败: %w", err)
	}

	slog.Info("设置头像成功", "user_id", userID)
	return nil
}

// Unregister 注销用户.
func (s *UserService) Unregister(ctx context.Context, userID, email, password, verifyCode string) error {
	// 1. 验证码校验
	storedCode, err := s.verifyCodeRepo.GetVerifyCode(ctx, email)
	if err != nil {
		return errors.New("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		return errors.New("验证码错误")
	}

	// 2. 查找用户并校验密码
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if err := CheckPassword(password, user.Password); err != nil {
		return errors.New("密码错误")
	}

	// 3. 删除用户
	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		slog.Error("注销用户失败", "user_id", userID, "error", err)
		return fmt.Errorf("注销用户失败: %w", err)
	}

	// 4. 清理该用户的所有会话
	if err := s.sessionRepo.InvalidateUserSessions(ctx, userID); err != nil {
		slog.Warn("注销时清理会话失败", "user_id", userID, "error", err)
	}

	// 5. 删除已使用的验证码
	if err := s.verifyCodeRepo.DeleteVerifyCode(ctx, email); err != nil {
		slog.Warn("删除验证码失败", "email", email, "error", err)
	}

	slog.Info("用户注销成功", "user_id", userID)
	return nil
}
