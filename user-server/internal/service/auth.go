package service

import (
	"context"
	"fmt"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
	"gorm.io/gorm"
)

// Register 注册用户
func (s *Service) Register(ctx context.Context, name, email, password, verifyCode string) (*schema.User, error) {
	// 1. 参数校验
	if err := validateRegisterParams(name, email, password, verifyCode); err != nil {
		s.Logger.Logger.Warn("注册参数校验失败", "email", email, "error", err)
		return nil, err
	}

	// 2. 验证码校验
	storedCode, err := s.Repository.GetVerifyCode(ctx, email)
	if err != nil {
		s.Logger.Logger.Error("获取验证码失败", "email", email, "error", err)
		return nil, fmt.Errorf("验证码已过期或不存在")
	}
	if storedCode != verifyCode {
		s.Logger.Logger.Warn("验证码错误", "email", email)
		return nil, fmt.Errorf("验证码错误")
	}

	// 3. 检查邮箱是否已被注册
	existingUser, err := s.Repository.GetUserByEmail(ctx, email)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.Logger.Logger.Error("检查邮箱是否已注册失败", "email", email, "error", err)
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existingUser != nil {
		s.Logger.Logger.Warn("邮箱已被注册", "email", email)
		return nil, fmt.Errorf("该邮箱已被注册")
	}

	// 4. 密码加密
	hashedPassword, err := HashPassword(password)
	if err != nil {
		s.Logger.Logger.Error("密码加密失败", "email", email, "error", err)
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 5. 创建用户
	user := &schema.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}
	if err := s.Repository.CreateUser(user); err != nil {
		s.Logger.Logger.Error("创建用户失败", "email", email, "error", err)
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 6. 删除已使用的验证码
	if err := s.Repository.DeleteVerifyCode(ctx, email); err != nil {
		s.Logger.Logger.Warn("删除验证码失败", "email", email, "error", err)
	}

	s.Logger.Logger.Info("用户注册成功", "user_id", user.ID, "email", email)
	return user, nil
}

// Login 用户登录
func (s *Service) Login(ctx context.Context, email, password, clientIP string) (string, string, error) {
	// TODO: implement me
	return "", "", nil
}

// Logout 用户退出登录
func (s *Service) Logout(ctx context.Context, refreshToken, clientIP string) error {
	// TODO: implement me
	return nil
}

// Refresh 刷新登录凭证
func (s *Service) Refresh(ctx context.Context, refreshToken, clientIP string) (string, string, error) {
	// TODO: implement me
	return "", "", nil
}
