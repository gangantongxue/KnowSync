package schema

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Repo 知识库表.
type Repo struct {
	ID           string         `gorm:"primaryKey;type:char(20)" json:"id"`
	OwnerID      string         `gorm:"column:owner_id;type:varchar(20);not null;index" json:"owner_id"`
	Name         string         `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Visibility   string         `gorm:"column:visibility;type:varchar(20);not null;default:PRIVATE" json:"visibility"`
	Description  string         `gorm:"column:description;type:varchar(500)" json:"description"`
	ArticleCount int64          `gorm:"column:article_count;not null;default:0" json:"article_count"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// TableName 返回知识库表名.
func (r *Repo) TableName() string {
	return "repo"
}

// BeforeCreate GORM 钩子，在创建前自动生成 ID.
func (r *Repo) BeforeCreate(_ *gorm.DB) error {
	if r.ID == "" {
		r.ID = xid.New().String()
	}
	return nil
}
