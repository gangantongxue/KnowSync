package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// GetUser 获取用户信息
func (s *Service) GetUser(ctx context.Context, userID string) (*schema.User, error) {
	user, err := s.Repository.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		s.Logger.Logger.Error("获取用户信息失败", "user_id", userID, "error", err)
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	return user, nil
}

// UpdateUserInfo 更新用户信息
func (s *Service) UpdateUserInfo(ctx context.Context, userID, name, email, avatar string) (*schema.User, error) {
	user, err := s.Repository.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
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

	if err := s.Repository.UpdateUser(ctx, user); err != nil {
		s.Logger.Logger.Error("更新用户信息失败", "user_id", userID, "error", err)
		return nil, fmt.Errorf("更新用户信息失败: %w", err)
	}

	s.Logger.Logger.Info("更新用户信息成功", "user_id", userID)
	return user, nil
}

// SetAvatar 设置用户头像
func (s *Service) SetAvatar(ctx context.Context, userID, avatar string) error {
	user, err := s.Repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	user.Avatar = avatar
	if err := s.Repository.UpdateUser(ctx, user); err != nil {
		s.Logger.Logger.Error("设置头像失败", "user_id", userID, "error", err)
		return fmt.Errorf("设置头像失败: %w", err)
	}

	s.Logger.Logger.Info("设置头像成功", "user_id", userID)
	return nil
}

// Unregister 注销用户
func (s *Service) Unregister(ctx context.Context, userID, email, password, verifyCode string) error {
	// 1. 验证码校验
	storedCode, err := s.Repository.GetVerifyCode(ctx, email)
	if err != nil {
		return fmt.Errorf("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		return fmt.Errorf("验证码错误")
	}

	// 2. 查找用户并校验密码
	user, err := s.Repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if err := CheckPassword(password, user.Password); err != nil {
		return fmt.Errorf("密码错误")
	}

	// 3. 删除用户
	if err := s.Repository.DeleteUser(ctx, userID); err != nil {
		s.Logger.Logger.Error("注销用户失败", "user_id", userID, "error", err)
		return fmt.Errorf("注销用户失败: %w", err)
	}

	// 4. 清理该用户的所有会话
	if err := s.Repository.InvalidateUserSessions(ctx, userID); err != nil {
		s.Logger.Logger.Warn("注销时清理会话失败", "user_id", userID, "error", err)
	}

	// 5. 删除已使用的验证码
	if err := s.Repository.DeleteVerifyCode(ctx, email); err != nil {
		s.Logger.Logger.Warn("删除验证码失败", "email", email, "error", err)
	}

	s.Logger.Logger.Info("用户注销成功", "user_id", userID)
	return nil
}
