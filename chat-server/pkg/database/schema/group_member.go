package schema

// GroupMember 群组成员表
type GroupMember struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	GroupID       uint64 `gorm:"column:group_id;type:bigint unsigned;not null;uniqueIndex:uk_group_user,priority:1" json:"group_id"`
	UserID        uint64 `gorm:"column:user_id;type:bigint unsigned;not null;uniqueIndex:uk_group_user,priority:2;index:idx_user" json:"user_id"`
	Role          string `gorm:"column:role;type:enum('owner','admin','member');not null;default:member" json:"role"`
	LastReadSeqID uint64 `gorm:"column:last_read_seq_id;type:bigint unsigned;not null;default:0" json:"last_read_seq_id"`
	Pinned        int8   `gorm:"column:pinned;type:tinyint(1);not null;default:0" json:"pinned"`
	JoinedAt      int64  `gorm:"column:joined_at;type:bigint;not null" json:"joined_at"`
}

func (g *GroupMember) TableName() string {
	return "group_members"
}
