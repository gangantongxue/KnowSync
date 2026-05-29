package repository

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

// ChatMessage 消息表
type ChatMessage struct {
	ID        string    `gorm:"primaryKey;type:char(20)" json:"id"`
	SessionID string    `gorm:"column:session_id;type:char(20);not null;index:idx_session_id" json:"session_id"`
	Role      string    `gorm:"column:role;type:varchar(16);not null" json:"role"` // "user" / "assistant"
	Content   string    `gorm:"column:content;type:longtext" json:"content"`       // 正式回答
	Thinking  string    `gorm:"column:thinking;type:longtext" json:"thinking"`     // 思考过程
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_session_id" json:"created_at"`
}

func (m *ChatMessage) TableName() string {
	return "chat_message"
}

// BeforeCreate 在创建前调用
func (m *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = xid.New().String()
	}
	return nil
}

// CreateMessage 创建消息
func (r *Repository) CreateMessage(msg *ChatMessage) error {
	return r.DB.Create(msg).Error
}

// ListMessages 获取会话的消息列表（按创建时间正序游标分页）
// cursor 为 created_at 时间戳（秒），首次传 0 表示从头开始
// 返回消息列表及是否有更多数据
func (r *Repository) ListMessages(sessionID string, cursor int64, limit int) ([]ChatMessage, bool, error) {
	var messages []ChatMessage

	query := r.DB.Model(&ChatMessage{}).Where("session_id = ?", sessionID)

	if cursor > 0 {
		query = query.Where("created_at > ?", time.Unix(cursor, 0))
	}

	if err := query.Order("created_at ASC").Limit(limit + 1).Find(&messages).Error; err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	return messages, hasMore, nil
}

// GetSessionMessages 获取会话所有消息（按创建时间正序，用于构建 LLM 上下文）
func (r *Repository) GetSessionMessages(sessionID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	err := r.DB.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}
