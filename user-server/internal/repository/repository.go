// Package repository provides data access layer for the user-server.
package repository

import (
	"github.com/gangantongxue/knowsync/user-server/pkg/database"
	"github.com/gangantongxue/knowsync/user-server/pkg/redis"
)

// Repository 仓库层.
type Repository struct {
	Database *database.Database
	Redis    *redis.Redis
}

// NewRepository 创建仓库层.
func NewRepository(db *database.Database, r *redis.Redis) (*Repository, error) {
	return &Repository{Database: db, Redis: r}, nil
}
