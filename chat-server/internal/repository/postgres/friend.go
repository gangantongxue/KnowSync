package postgres

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// FriendRepo 好友仓库实现.
type FriendRepo struct {
	db *gorm.DB
}

// NewFriendRepo 创建好友仓库.
func NewFriendRepo(db *gorm.DB) *FriendRepo {
	return &FriendRepo{db: db}
}

// AddFriend 添加好友.
func (r *FriendRepo) AddFriend(ctx context.Context, friend *schema.Friend) error {
	return r.db.WithContext(ctx).Create(friend).Error
}

// RemoveFriend 移除好友.
func (r *FriendRepo) RemoveFriend(ctx context.Context, userID, friendID string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&schema.Friend{}).Error
}

// IsFriend 检查是否是好友.
func (r *FriendRepo) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&schema.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error
	return count > 0, err
}

// ListFriends 列出好友.
func (r *FriendRepo) ListFriends(ctx context.Context, userID string) ([]*schema.Friend, error) {
	var friends []*schema.Friend
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}
