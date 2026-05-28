package schema

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

// Node 节点表，统一表示文件夹和文章，通过 parent_id 构建树形结构，使用 "__ROOT__" 表示根节点
type Node struct {
	ID        string         `gorm:"primaryKey;type:char(20)" json:"id"`
	RepoID    string         `gorm:"column:repo_id;type:char(20);not null;uniqueIndex:idx_repo_parent_name,priority:1" json:"repo_id"`
	ParentID  string         `gorm:"column:parent_id;type:char(20);not null;default:__ROOT__;uniqueIndex:idx_repo_parent_name,priority:2" json:"parent_id"`
	Name      string         `gorm:"column:name;type:varchar(255);not null;uniqueIndex:idx_repo_parent_name,priority:3" json:"name"`
	Type      string         `gorm:"column:type;type:varchar(10);not null" json:"type"`
	FilePath  string         `gorm:"column:file_path;type:varchar(500)" json:"file_path"`
	Size      int64          `gorm:"column:size;not null;default:0" json:"size"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;uniqueIndex:idx_repo_parent_name,priority:4" json:"deleted_at"`
}

func (n *Node) TableName() string {
	return "node"
}

func (n *Node) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = xid.New().String()
	}
	return nil
}
