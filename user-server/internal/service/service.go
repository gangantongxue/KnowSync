package service

import (
	"crypto/rsa"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/user-server/internal/repository"
	"github.com/gangantongxue/knowsync/user-server/pkg/auth"
	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
	"github.com/gangantongxue/knowsync/user-server/pkg/mail"
)

// Service 业务服务层.
type Service struct {
	Cfg        *config.Config
	Repository *repository.Repository
	Mailer     *mail.Mailer
	Logger     *logger.Logger
	privateKey *rsa.PrivateKey
}

// NewService 创建业务服务层.
func NewService(cfg *config.Config, repository *repository.Repository, mailer *mail.Mailer, l *logger.Logger) (*Service, error) {
	privateKey, err := auth.LoadPrivateKeyFromFile(cfg.Auth.RSAPrivateKeyPath)
	if err != nil {
		slog.Error("加载 RSA 私钥失败", "error", err)
		return nil, fmt.Errorf("加载 RSA 私钥失败: %w", err)
	}

	return &Service{
		Cfg:        cfg,
		Repository: repository,
		Mailer:     mailer,
		Logger:     l,
		privateKey: privateKey,
	}, nil
}
