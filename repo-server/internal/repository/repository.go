package repository

import (
	"github.com/gangantongxue/knowsync/repo-server/pkg/database"
)

// Repository 仓库层，封装所有数据库操作，遵循无外键无 JOIN 规范
type Repository struct {
	Database *database.Database
}

func NewRepository(db *database.Database) (*Repository, error) {
	return &Repository{Database: db}, nil
}
