package schema

// FriendRequest 好友申请表
type FriendRequest struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	SenderID   uint64 `gorm:"column:sender_id;type:bigint unsigned;not null;index:idx_sender" json:"sender_id"`
	ReceiverID uint64 `gorm:"column:receiver_id;type:bigint unsigned;not null;index:idx_receiver_status,priority:1" json:"receiver_id"`
	Status     string `gorm:"column:status;type:enum('pending','accepted','rejected');not null;default:pending;index:idx_receiver_status,priority:2" json:"status"`
	Remark     string `gorm:"column:remark;type:varchar(100);default:''" json:"remark"`
	CreatedAt  int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (f *FriendRequest) TableName() string {
	return "friend_requests"
}
