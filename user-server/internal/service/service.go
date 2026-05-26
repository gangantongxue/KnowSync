package service

import "github.com/gangantongxue/knowsync/user-server/internal/repository"

// Service 服务层
type Service struct {
	Repository *repository.Repository
}

// NewService 创建服务层
func NewService(repository *repository.Repository) (*Service, error) {
	return &Service{Repository: repository}, nil
}
