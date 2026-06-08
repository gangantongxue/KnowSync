package postgres

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// MessageRepo 消息仓库实现.
type MessageRepo struct {
	db *gorm.DB
}

// NewMessageRepo 创建消息仓库.
func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// CreateMessage 创建消息.
func (r *MessageRepo) CreateMessage(ctx context.Context, msg *schema.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// GetMessage 获取消息.
func (r *MessageRepo) GetMessage(ctx context.Context, id string) (*schema.Message, error) {
	var msg schema.Message
	if err := r.db.WithContext(ctx).First(&msg, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

// ListMessages 列出消息.
func (r *MessageRepo) ListMessages(ctx context.Context, conversationID string, page, pageSize int) ([]*schema.Message, int64, error) {
	var msgs []*schema.Message
	var total int64

	query := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}

	return msgs, total, nil
}

// UpdateMessageStatus 更新消息状态.
func (r *MessageRepo) UpdateMessageStatus(ctx context.Context, msgID, status string) error {
	return r.db.WithContext(ctx).Model(&schema.Message{}).
		Where("id = ?", msgID).
		Update("status", status).Error
}

// DeleteMessage 删除消息.
func (r *MessageRepo) DeleteMessage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&schema.Message{}, "id = ?", id).Error
}
