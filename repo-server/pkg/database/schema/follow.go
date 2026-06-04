package schema

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Follow 关注表，记录用户关注的知识库.
type Follow struct {
	ID        string    `gorm:"primaryKey;type:char(20)" json:"id"`
	UserID    string    `gorm:"column:user_id;type:varchar(20);not null;uniqueIndex:idx_user_repo,priority:1" json:"user_id"`
	RepoID    string    `gorm:"column:repo_id;type:char(20);not null;uniqueIndex:idx_user_repo,priority:2" json:"repo_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 返回关注表名.
func (f *Follow) TableName() string {
	return "follow"
}

// BeforeCreate GORM 钩子，在创建前自动生成 ID.
func (f *Follow) BeforeCreate(_ *gorm.DB) error {
	if f.ID == "" {
		f.ID = xid.New().String()
	}
	return nil
}
