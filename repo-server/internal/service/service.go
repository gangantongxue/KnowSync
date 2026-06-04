package service

import (
	"errors"

	"github.com/gangantongxue/knowsync/repo-server/internal/repository"
)

var (
	// ErrPermissionDenied 权限不足错误.
	ErrPermissionDenied = errors.New("权限不足")
	// ErrNotFound 资源不存在错误.
	ErrNotFound = errors.New("资源不存在")
	// ErrAlreadyExists 资源已存在错误.
	ErrAlreadyExists = errors.New("已存在同名资源")
)

// Service 服务层，封装业务逻辑和权限校验.
type Service struct {
	Repository *repository.Repository
}

// NewService 创建服务层实例.
func NewService(repo *repository.Repository) (*Service, error) {
	return &Service{Repository: repo}, nil
}
