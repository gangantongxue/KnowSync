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

// ListSessions 列出用户的会话（按更新时间倒序游标分页）
// cursor 为 updated_at 时间戳（秒），首次传 0 表示从头开始
// 返回会话列表及是否有更多数据
func (r *Repository) ListSessions(userID string, cursor int64, limit int) ([]ChatSession, bool, error) {
	var sessions []ChatSession

	query := r.DB.Model(&ChatSession{}).Where("user_id = ?", userID)

	if cursor > 0 {
		query = query.Where("updated_at < ?", time.Unix(cursor, 0))
	}

	if err := query.Order("updated_at DESC").Limit(limit + 1).Find(&sessions).Error; err != nil {
		return nil, false, err
	}

	hasMore := len(sessions) > limit
	if hasMore {
		sessions = sessions[:limit]
	}

	return sessions, hasMore, nil
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
