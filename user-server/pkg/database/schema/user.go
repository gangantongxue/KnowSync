package schema

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID       string `gorm:"primaryKey,type:char(20)" json:"id"`
	Name     string `gorm:"column:name,type:varchar(64), not null" json:"name"`
	Email    string `gorm:"column:email,type:varchar(254), not null" json:"email"`
	Password string `gorm:"column:password,type:varchar(255), not null" json:"password"`
	Avatar   string `gorm:"column:avatar,type:varchar(255)" json:"avatar"`

	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

func (u *User) TableName() string {
	return "user"
}

// BeforeCreate 在创建用户前调用
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = xid.New().String()
	}
	return nil
}
