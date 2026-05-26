package service

import (
	"context"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// Register 注册用户
func (s *Service) Register(ctx context.Context, name, email, password, verifyCode string) (*schema.User, error) {
	// TODO: implement me
	return nil, nil
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
