package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// ConversationRepository 会话仓库
type ConversationRepository struct {
	DB *gorm.DB
}

// NewConversationRepository 创建会话仓库
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{DB: db}
}

// GetFriendConversations 获取用户的所有好友会话
func (r *ConversationRepository) GetFriendConversations(ctx context.Context, userID string) ([]schema.Friend, error) {
	var friends []schema.Friend
	if err := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}

// GetGroupConversations 获取用户的所有群组会话（包含群组信息和成员角色）
func (r *ConversationRepository) GetGroupConversations(ctx context.Context, userID string) ([]GroupConversationInfo, error) {
	var results []GroupConversationInfo
	if err := r.DB.WithContext(ctx).
		Table("groups").
		Select("groups.*, group_members.role, group_members.last_read_seq_id, group_members.pinned").
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GroupConversationInfo 群组会话信息（包含成员角色和已读进度）
type GroupConversationInfo struct {
	schema.Group
	Role          string `gorm:"column:role"`
	LastReadSeqID uint64 `gorm:"column:last_read_seq_id"`
	Pinned        int8   `gorm:"column:pinned"`
}

// UpdateFriendLastReadSeqID 更新好友最后读取的 seq_id
func (r *ConversationRepository) UpdateFriendLastReadSeqID(ctx context.Context, userID, friendID string, seqID uint64) error {
	return r.DB.WithContext(ctx).
		Model(&schema.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("last_read_seq_id", seqID).Error
}

// UpdateGroupMemberLastReadSeqID 更新群成员最后读取的 seq_id
func (r *ConversationRepository) UpdateGroupMemberLastReadSeqID(ctx context.Context, groupID, userID string, seqID uint64) error {
	return r.DB.WithContext(ctx).
		Model(&schema.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("last_read_seq_id", seqID).Error
}

// ToggleFriendPin 切换好友置顶
func (r *ConversationRepository) ToggleFriendPin(ctx context.Context, userID, friendID string) (bool, error) {
	var friend schema.Friend
	if err := r.DB.WithContext(ctx).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		First(&friend).Error; err != nil {
		return false, err
	}
	newPinned := int8(1)
	if friend.Pinned == 1 {
		newPinned = 0
	}
	if err := r.DB.WithContext(ctx).
		Model(&schema.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("pinned", newPinned).Error; err != nil {
		return false, err
	}
	return newPinned == 1, nil
}

// ToggleGroupPin 切换群组置顶
func (r *ConversationRepository) ToggleGroupPin(ctx context.Context, groupID, userID string) (bool, error) {
	var member schema.GroupMember
	if err := r.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error; err != nil {
		return false, err
	}
	newPinned := int8(1)
	if member.Pinned == 1 {
		newPinned = 0
	}
	if err := r.DB.WithContext(ctx).
		Model(&schema.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("pinned", newPinned).Error; err != nil {
		return false, err
	}
	return newPinned == 1, nil
}

// DeleteFriendRelation 删除好友关系（双向）
func (r *ConversationRepository) DeleteFriendRelation(ctx context.Context, userID, friendID string) error {
	return r.DB.WithContext(ctx).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userID, friendID, friendID, userID).
		Delete(&schema.Friend{}).Error
}
