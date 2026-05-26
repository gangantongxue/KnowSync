package repository

import (
	"context"
	"time"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// CreateSession 创建用户会话
func (r *Repository) CreateSession(ctx context.Context, session *schema.UserSession) error {
	return r.Database.DB.WithContext(ctx).Create(session).Error
}

// GetSessionByRefreshToken 根据 refresh token 获取会话
func (r *Repository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*schema.UserSession, error) {
	var session schema.UserSession
	if err := r.Database.DB.WithContext(ctx).
		Where("refresh_token_hash = ?", refreshToken).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession 更新用户会话
func (r *Repository) UpdateSession(ctx context.Context, session *schema.UserSession) error {
	return r.Database.DB.WithContext(ctx).Model(session).Updates(session).Error
}

// InvalidateUserSessions 将用户的所有活跃会话标记为已退出
func (r *Repository) InvalidateUserSessions(ctx context.Context, userID string) error {
	now := time.Now()
	return r.Database.DB.WithContext(ctx).
		Model(&schema.UserSession{}).
		Where("user_id = ? AND logout_at IS NULL", userID).
		Update("logout_at", now).Error
}
