package schema

// User 用户表（用于搜索用户功能）.
type User struct {
	ID       string `gorm:"primaryKey;type:char(20)" json:"id"`
	Name     string `gorm:"column:name;type:varchar(50);not null" json:"name"`
	Email    string `gorm:"column:email;type:varchar(100);not null;uniqueIndex" json:"email"`
	Avatar   string `gorm:"column:avatar;type:varchar(255);default:''" json:"avatar"`
	Password string `gorm:"column:password;type:varchar(255);not null" json:"-"`
}

// TableName 返回用户表名.
func (u *User) TableName() string {
	return "users"
}
