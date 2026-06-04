package schema

import (
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Collaborator 协作者表，记录知识库的用户权限
type Collaborator struct {
	ID     string `gorm:"primaryKey;type:char(20)" json:"id"`
	RepoID string `gorm:"column:repo_id;type:char(20);not null;uniqueIndex:idx_repo_user,priority:1" json:"repo_id"`
	UserID string `gorm:"column:user_id;type:varchar(20);not null;uniqueIndex:idx_repo_user,priority:2" json:"user_id"`
	Role   string `gorm:"column:role;type:varchar(20);not null" json:"role"`
}

func (c *Collaborator) TableName() string {
	return "collaborator"
}

func (c *Collaborator) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = xid.New().String()
	}
	return nil
}
