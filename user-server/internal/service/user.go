package service

import (
	"context"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

func (s *Service) Register(ctx context.Context, name, email, password, verifyCode string) (*schema.User, error) {
	return nil, nil
}