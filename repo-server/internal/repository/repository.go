// Package repository 提供数据访问层实现，封装所有数据库操作.
package repository

import (
	"github.com/gangantongxue/knowsync/repo-server/pkg/database"
)

// Repository 仓库层，封装所有数据库操作，遵循无外键无 JOIN 规范.
type Repository struct {
	Database *database.Database
}

// NewRepository 创建 Repository 实例.
func NewRepository(db *database.Database) (*Repository, error) {
	return &Repository{Database: db}, nil
}
