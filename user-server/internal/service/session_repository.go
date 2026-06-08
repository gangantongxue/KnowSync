package service

import (
	"context"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// SessionRepository 会话数据访问接口（消费者定义）.
type SessionRepository interface {
	CreateSession(ctx context.Context, session *schema.UserSession) error
	GetSessionByRefreshToken(ctx context.Context, tokenHash string) (*schema.UserSession, error)
	UpdateSession(ctx context.Context, session *schema.UserSession) error
	InvalidateUserSessions(ctx context.Context, userID string) error
}
