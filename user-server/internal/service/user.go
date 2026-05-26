package service

import (
	"context"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// GetUser 获取用户信息
func (s *Service) GetUser(ctx context.Context, userID string) (*schema.User, error) {
	// TODO: implement me
	return nil, nil
}

// UpdateUserInfo 更新用户信息
func (s *Service) UpdateUserInfo(ctx context.Context, userID, name, email, avatar string) (*schema.User, error) {
	// TODO: implement me
	return nil, nil
}

// SetAvatar 设置用户头像
func (s *Service) SetAvatar(ctx context.Context, userID, avatar string) error {
	// TODO: implement me
	return nil
}

// Unregister 注销用户
func (s *Service) Unregister(ctx context.Context, userID, email, password, verifyCode string) error {
	// TODO: implement me
	return nil
}
