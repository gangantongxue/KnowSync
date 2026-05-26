package schema

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID       string `gorm:"primaryKey;type:char(20)" json:"id"`
	Name     string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Email    string `gorm:"column:email;type:varchar(254);not null;uniqueIndex" json:"email"`
	Password string `gorm:"column:password;type:varchar(255);not null" json:"password"`
	Avatar   string `gorm:"column:avatar;type:varchar(255)" json:"avatar"`

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
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

// UserSession 用户会话表
type UserSession struct {
	ID           string         `gorm:"primaryKey;type:char(20)" json:"id"`
	UserID       string         `gorm:"column:user_id;type:char(20);not null;index:user_id_index;uniqueIndex:user_id_refresh_token_index,priority:1" json:"user_id"`
	ClientIP     string         `gorm:"column:client_ip;type:varchar(64);not null" json:"client_ip"`
	RefreshToken string         `gorm:"column:refresh_token;type:varchar(255);not null;uniqueIndex:user_id_refresh_token_index,priority:2" json:"refresh_token"`
	ExpireAt     time.Time      `gorm:"column:expire_at;not null" json:"expire_at"`
	LoginAt      time.Time      `gorm:"column:login_at;not null" json:"login_at"`
	LogoutAt     *time.Time     `gorm:"column:logout_at" json:"logout_at"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (u *UserSession) TableName() string {
	return "user_session"
}

// BeforeCreate 在创建用户会话前调用
func (u *UserSession) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = xid.New().String()
	}
	return nil
}
