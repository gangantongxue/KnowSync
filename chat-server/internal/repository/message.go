package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// MessageRepository 消息仓库.
type MessageRepository struct {
	DB *gorm.DB
}

// NewMessageRepository 创建消息仓库.
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}

// CreateMessage 创建消息.
func (r *MessageRepository) CreateMessage(ctx context.Context, msg *schema.Message) error {
	return r.DB.WithContext(ctx).Create(msg).Error
}

// GetMessagesByConversation 获取会话消息列表（游标分页，返回按 seq_id ASC 排序）.
func (r *MessageRepository) GetMessagesByConversation(ctx context.Context, conversationType, conversationID string, beforeSeqID uint64, limit int) ([]schema.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var messages []schema.Message
	db := r.DB.WithContext(ctx).
		Where("conversation_type = ? AND conversation_id = ?", conversationType, conversationID)
	if beforeSeqID > 0 {
		db = db.Where("seq_id < ?", beforeSeqID)
	}
	// 查询最新的 limit 条消息（DESC），再反转为 ASC 供前端按时间正序展示
	if err := db.Order("seq_id DESC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// GetMessageByID 根据ID获取消息.
func (r *MessageRepository) GetMessageByID(ctx context.Context, messageID string) (*schema.Message, error) {
	var msg schema.Message
	if err := r.DB.WithContext(ctx).Where("id = ?", messageID).First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetLastMessageByConversation 获取会话最新消息.
func (r *MessageRepository) GetLastMessageByConversation(ctx context.Context, conversationType, conversationID string) (*schema.Message, error) {
	var msg schema.Message
	if err := r.DB.WithContext(ctx).
		Where("conversation_type = ? AND conversation_id = ?", conversationType, conversationID).
		Order("seq_id DESC").
		First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// CountMessagesAfter 统计某个seq_id之后的消息数.
func (r *MessageRepository) CountMessagesAfter(ctx context.Context, conversationType, conversationID string, afterSeqID uint64) (int64, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&schema.Message{}).
		Where("conversation_type = ? AND conversation_id = ? AND seq_id > ?", conversationType, conversationID, afterSeqID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// RecallMessage 撤回消息（设置状态为 recalled）.
func (r *MessageRepository) RecallMessage(ctx context.Context, messageID, senderID string) error {
	return r.DB.WithContext(ctx).
		Model(&schema.Message{}).
		Where("id = ? AND sender_id = ?", messageID, senderID).
		Update("status", "recalled").Error
}

// GetMessagesByIDs 批量获取消息.
func (r *MessageRepository) GetMessagesByIDs(ctx context.Context, messageIDs []string) ([]schema.Message, error) {
	var messages []schema.Message
	if err := r.DB.WithContext(ctx).
		Where("id IN ?", messageIDs).
		Order("seq_id ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetMaxSeqID 获取会话最大seq_id.
func (r *MessageRepository) GetMaxSeqID(ctx context.Context, conversationType, conversationID string) (uint64, error) {
	var maxSeqID uint64
	if err := r.DB.WithContext(ctx).
		Model(&schema.Message{}).
		Select("COALESCE(MAX(seq_id), 0)").
		Where("conversation_type = ? AND conversation_id = ?", conversationType, conversationID).
		Scan(&maxSeqID).Error; err != nil {
		return 0, err
	}
	return maxSeqID, nil
}
