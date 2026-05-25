package repository

import "github.com/gangantongxue/knowsync/user-server/pkg/database"

// Repository 仓库层
type Repository struct {
	Database *database.Database
}

// NewRepository 创建仓库层
func NewRepository(db *database.Database) (*Repository, error) {
	return &Repository{Database: db}, nil
}