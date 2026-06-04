package service

import (
	"github.com/gangantongxue/knowsync/chat-server/internal/repository"
	"github.com/gangantongxue/knowsync/chat-server/internal/ws"
)

// Service 业务服务，聚合所有业务逻辑
type Service struct {
	Repo *repository.Repository
	Hub  *ws.Hub
}

// NewService 创建业务服务
func NewService(repo *repository.Repository, hub *ws.Hub) *Service {
	return &Service{Repo: repo, Hub: hub}
}
