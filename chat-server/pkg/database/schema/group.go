package schema

import (
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Group 群组表
type Group struct {
	ID        string `gorm:"primaryKey;type:char(20)" json:"id"`
	Name      string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Avatar    string `gorm:"column:avatar;type:varchar(500);default:''" json:"avatar"`
	OwnerID   string `gorm:"column:owner_id;type:varchar(20);not null" json:"owner_id"`
	CreatedAt int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (g *Group) TableName() string {
	return "groups"
}

func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = xid.New().String()
	}
	return nil
}
