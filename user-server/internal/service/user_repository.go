package service

import (
	"context"

	"github.com/gangantongxue/knowsync/user-server/pkg/database/schema"
)

// UserRepository 用户数据访问接口（消费者定义）.
type UserRepository interface {
	GetUser(ctx context.Context, id string) (*schema.User, error)
	GetUserByEmail(ctx context.Context, email string) (*schema.User, error)
	CreateUser(ctx context.Context, user *schema.User) error
	UpdateUser(ctx context.Context, user *schema.User) error
	DeleteUser(ctx context.Context, id string) error
}
