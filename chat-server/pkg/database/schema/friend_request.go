package schema

import (
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// FriendRequest 好友申请表.
type FriendRequest struct {
	ID         string `gorm:"primaryKey;type:char(20)" json:"id"`
	SenderID   string `gorm:"column:sender_id;type:varchar(20);not null;index:idx_sender" json:"sender_id"`
	ReceiverID string `gorm:"column:receiver_id;type:varchar(20);not null;index:idx_receiver_status,priority:1" json:"receiver_id"`
	Status     string `gorm:"column:status;type:enum('pending','accepted','rejected');not null;default:pending;index:idx_receiver_status,priority:2" json:"status"`
	Remark     string `gorm:"column:remark;type:varchar(100);default:''" json:"remark"`
	CreatedAt  int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

// TableName 返回好友申请表名.
func (f *FriendRequest) TableName() string {
	return "friend_requests"
}

// BeforeCreate GORM 创建前钩子，自动生成 ID.
func (f *FriendRequest) BeforeCreate(_ *gorm.DB) error {
	if f.ID == "" {
		f.ID = xid.New().String()
	}
	return nil
}
