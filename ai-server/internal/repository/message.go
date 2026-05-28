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

// ListMessages 获取会话的消息列表（按创建时间正序）
func (r *Repository) ListMessages(sessionID string, page, pageSize int) ([]ChatMessage, int64, error) {
	var messages []ChatMessage
	var total int64

	query := r.DB.Model(&ChatMessage{}).Where("session_id = ?", sessionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetSessionMessages 获取会话所有消息（按创建时间正序，用于构建 LLM 上下文）
func (r *Repository) GetSessionMessages(sessionID string) ([]ChatMessage, error) {
	var messages []ChatMessage
	err := r.DB.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}
