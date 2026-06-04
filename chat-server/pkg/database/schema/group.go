package schema

// Group 群组表
type Group struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	Name      string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Avatar    string `gorm:"column:avatar;type:varchar(500);default:''" json:"avatar"`
	OwnerID   uint64 `gorm:"column:owner_id;type:bigint unsigned;not null" json:"owner_id"`
	CreatedAt int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (g *Group) TableName() string {
	return "groups"
}
