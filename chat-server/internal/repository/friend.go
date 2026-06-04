package repository

import (
	"context"
	"time"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// FriendRepository 好友关系仓库
type FriendRepository struct {
	DB *gorm.DB
}

// NewFriendRepository 创建好友关系仓库
func NewFriendRepository(db *gorm.DB) *FriendRepository {
	return &FriendRepository{DB: db}
}

// CreateFriendRequest 创建好友申请
func (r *FriendRepository) CreateFriendRequest(ctx context.Context, request *schema.FriendRequest) error {
	return r.DB.WithContext(ctx).Create(request).Error
}

// GetFriendRequestByID 根据ID获取好友申请
func (r *FriendRepository) GetFriendRequestByID(ctx context.Context, id string) (*schema.FriendRequest, error) {
	var request schema.FriendRequest
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

// GetPendingFriendRequest 检查是否存在待处理的好友申请
func (r *FriendRepository) GetPendingFriendRequest(ctx context.Context, senderID, receiverID string) (*schema.FriendRequest, error) {
	var request schema.FriendRequest
	if err := r.DB.WithContext(ctx).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			senderID, receiverID, receiverID, senderID).
		Where("status = ?", "pending").
		First(&request).Error; err != nil {
		return nil, err
	}
	return &request, nil
}

// GetFriendRequestsByReceiver 获取收到的好友申请列表
func (r *FriendRepository) GetFriendRequestsByReceiver(ctx context.Context, receiverID string) ([]schema.FriendRequest, error) {
	var requests []schema.FriendRequest
	if err := r.DB.WithContext(ctx).
		Where("receiver_id = ? AND status = ?", receiverID, "pending").
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// GetFriendRequestsBySender 获取发送的好友申请列表
func (r *FriendRepository) GetFriendRequestsBySender(ctx context.Context, senderID string) ([]schema.FriendRequest, error) {
	var requests []schema.FriendRequest
	if err := r.DB.WithContext(ctx).
		Where("sender_id = ?", senderID).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// UpdateFriendRequestStatus 更新好友申请状态
func (r *FriendRepository) UpdateFriendRequestStatus(ctx context.Context, id string, status string) error {
	return r.DB.WithContext(ctx).
		Model(&schema.FriendRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().Unix(),
		}).Error
}

// CreateFriend 创建好友关系
func (r *FriendRepository) CreateFriend(ctx context.Context, friend *schema.Friend) error {
	return r.DB.WithContext(ctx).Create(friend).Error
}

// GetFriend 获取好友关系
func (r *FriendRepository) GetFriend(ctx context.Context, userID, friendID string) (*schema.Friend, error) {
	var friend schema.Friend
	if err := r.DB.WithContext(ctx).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		First(&friend).Error; err != nil {
		return nil, err
	}
	return &friend, nil
}

// GetFriendList 获取好友列表
func (r *FriendRepository) GetFriendList(ctx context.Context, userID string, query string) ([]schema.Friend, error) {
	var friends []schema.Friend
	db := r.DB.WithContext(ctx).Where("user_id = ?", userID)

	if query != "" {
		db = db.Where("remark LIKE ?", "%"+query+"%")
	}

	if err := db.Order("last_message_at DESC").Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}

// DeleteFriend 删除好友关系
func (r *FriendRepository) DeleteFriend(ctx context.Context, userID, friendID string) error {
	return r.DB.WithContext(ctx).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userID, friendID, friendID, userID).
		Delete(&schema.Friend{}).Error
}

// UpdateFriendRemark 更新好友备注
func (r *FriendRepository) UpdateFriendRemark(ctx context.Context, userID, friendID, remark string) error {
	return r.DB.WithContext(ctx).
		Model(&schema.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("remark", remark).Error
}

// CheckFriendExists 检查好友关系是否存在
func (r *FriendRepository) CheckFriendExists(ctx context.Context, userID, friendID string) (bool, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&schema.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// SearchUserInfo 搜索用户信息（用于返回给调用方）
type SearchUserInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// UpdateFriendLastMessageAt 更新好友最后消息时间
func (r *FriendRepository) UpdateFriendLastMessageAt(ctx context.Context, userID, friendID string, lastMessageAt int64) error {
	return r.DB.WithContext(ctx).
		Model(&schema.Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("last_message_at", lastMessageAt).Error
}

// SearchUsers 搜索用户
func (r *FriendRepository) SearchUsers(ctx context.Context, query string) ([]SearchUserInfo, error) {
	var users []schema.User
	if err := r.DB.WithContext(ctx).
		Where("name LIKE ?", "%"+query+"%").
		Limit(20).
		Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为搜索结果格式
	var result []SearchUserInfo
	for _, u := range users {
		result = append(result, SearchUserInfo{
			ID:     u.ID,
			Name:   u.Name,
			Avatar: u.Avatar,
		})
	}
	return result, nil
}
