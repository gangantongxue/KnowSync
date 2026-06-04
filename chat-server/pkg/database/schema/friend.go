// Package schema 提供数据库模型定义.
package schema

import (
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Friend 好友关系表（双向各存一条，user_id 和 friend_id 互为好友）.
type Friend struct {
	ID            string `gorm:"primaryKey;type:char(20)" json:"id"`
	UserID        string `gorm:"column:user_id;type:varchar(20);not null;uniqueIndex:uk_user_friend,priority:1" json:"user_id"`
	FriendID      string `gorm:"column:friend_id;type:varchar(20);not null;uniqueIndex:uk_user_friend,priority:2" json:"friend_id"`
	Remark        string `gorm:"column:remark;type:varchar(100);default:''" json:"remark"`
	LastMessageAt int64  `gorm:"column:last_message_at;type:bigint;not null;default:0" json:"last_message_at"`
	LastReadSeqID uint64 `gorm:"column:last_read_seq_id;type:bigint unsigned;not null;default:0" json:"last_read_seq_id"`
	Pinned        int8   `gorm:"column:pinned;type:tinyint(1);not null;default:0" json:"pinned"`
	CreatedAt     int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
}

// TableName 返回好友关系表名.
func (f *Friend) TableName() string {
	return "friends"
}

// BeforeCreate GORM 创建前钩子，自动生成 ID.
func (f *Friend) BeforeCreate(_ *gorm.DB) error {
	if f.ID == "" {
		f.ID = xid.New().String()
	}
	return nil
}
