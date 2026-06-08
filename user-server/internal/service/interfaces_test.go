package service

import "github.com/gangantongxue/knowsync/user-server/internal/repository"

var (
	_ UserRepository       = (*repository.Repository)(nil)
	_ SessionRepository    = (*repository.Repository)(nil)
	_ VerifyCodeRepository = (*repository.Repository)(nil)
)
