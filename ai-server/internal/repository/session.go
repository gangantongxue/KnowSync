package repository

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

// ChatSession 会话表
type ChatSession struct {
	ID        string    `gorm:"primaryKey;type:char(20)" json:"id"`
	UserID    string    `gorm:"column:user_id;type:char(20);not null;index:idx_user_id" json:"user_id"`
	Title     string    `gorm:"column:title;type:varchar(255);not null;default:'新对话'" json:"title"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (s *ChatSession) TableName() string {
	return "chat_session"
}

// BeforeCreate 在创建前调用
func (s *ChatSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = xid.New().String()
	}
	return nil
}

// CreateSession 创建会话
func (r *Repository) CreateSession(session *ChatSession) error {
	return r.DB.Create(session).Error
}

// GetSession 获取会话
func (r *Repository) GetSession(sessionID string) (*ChatSession, error) {
	var session ChatSession
	err := r.DB.Where("id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// ListSessions 列出用户的会话（按更新时间倒序）
func (r *Repository) ListSessions(userID string, page, pageSize int) ([]ChatSession, int64, error) {
	var sessions []ChatSession
	var total int64

	query := r.DB.Model(&ChatSession{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// UpdateSessionTitle 更新会话标题
func (r *Repository) UpdateSessionTitle(sessionID, title string) error {
	return r.DB.Model(&ChatSession{}).Where("id = ?", sessionID).
		Update("title", title).Error
}

// DeleteSession 删除会话及其所有消息
func (r *Repository) DeleteSession(sessionID, userID string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&ChatMessage{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&ChatSession{}).Error
	})
}
