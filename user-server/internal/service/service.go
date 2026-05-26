package service

import (
	"github.com/gangantongxue/knowsync/user-server/internal/repository"
	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
	"github.com/gangantongxue/knowsync/user-server/pkg/mail"
)

// Service 服务层
type Service struct {
	Cfg        *config.Config
	Repository *repository.Repository
	Mailer     *mail.Mailer
	Logger     *logger.Logger
}

// NewService 创建服务层
func NewService(cfg *config.Config, repository *repository.Repository, mailer *mail.Mailer, l *logger.Logger) (*Service, error) {
	return &Service{
		Cfg:        cfg,
		Repository: repository,
		Mailer:     mailer,
		Logger:     l,
	}, nil
}
