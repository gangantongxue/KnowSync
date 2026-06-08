package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// CreateUser 创建用户.
func (r *Repository) CreateUser(ctx context.Context, user *schema.User) error {
	return r.Database.DB.WithContext(ctx).Create(user).Error
}

// GetUser 获取用户.
func (r *Repository) GetUser(ctx context.Context, id string) (*schema.User, error) {
	var user schema.User
	if err := r.Database.DB.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*schema.User, error) {
	var user schema.User
	if err := r.Database.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户.
func (r *Repository) UpdateUser(ctx context.Context, user *schema.User) error {
	return r.Database.DB.WithContext(ctx).Where("id = ?", user.ID).Updates(user).Error
}

// DeleteUser 删除用户.
func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	return r.Database.DB.WithContext(ctx).Where("id = ?", id).Delete(&schema.User{}).Error
}
