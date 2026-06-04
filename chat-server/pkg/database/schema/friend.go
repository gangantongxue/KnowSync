package schema

// Friend 好友关系表
type Friend struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	UserID        uint64 `gorm:"column:user_id;type:bigint unsigned;not null;uniqueIndex:uk_user_friend,priority:1" json:"user_id"`
	FriendID      uint64 `gorm:"column:friend_id;type:bigint unsigned;not null;uniqueIndex:uk_user_friend,priority:2" json:"friend_id"`
	Remark        string `gorm:"column:remark;type:varchar(100);default:''" json:"remark"`
	LastMessageAt int64  `gorm:"column:last_message_at;type:bigint;not null;default:0" json:"last_message_at"`
	LastReadSeqID uint64 `gorm:"column:last_read_seq_id;type:bigint unsigned;not null;default:0" json:"last_read_seq_id"`
	Pinned        int8   `gorm:"column:pinned;type:tinyint(1);not null;default:0" json:"pinned"`
	CreatedAt     int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
}

func (f *Friend) TableName() string {
	return "friends"
}
