package schema

import (
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// GroupMember 群组成员表
type GroupMember struct {
	ID            string `gorm:"primaryKey;type:char(20)" json:"id"`
	GroupID       string `gorm:"column:group_id;type:char(20);not null;uniqueIndex:uk_group_user,priority:1" json:"group_id"`
	UserID        string `gorm:"column:user_id;type:varchar(20);not null;uniqueIndex:uk_group_user,priority:2;index:idx_user" json:"user_id"`
	Role          string `gorm:"column:role;type:enum('owner','admin','member');not null;default:member" json:"role"`
	LastReadSeqID uint64 `gorm:"column:last_read_seq_id;type:bigint unsigned;not null;default:0" json:"last_read_seq_id"`
	Pinned        int8   `gorm:"column:pinned;type:tinyint(1);not null;default:0" json:"pinned"`
	JoinedAt      int64  `gorm:"column:joined_at;type:bigint;not null" json:"joined_at"`
}

func (g *GroupMember) TableName() string {
	return "group_members"
}

func (g *GroupMember) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = xid.New().String()
	}
	return nil
}
