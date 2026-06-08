package service

import "context"

// VerifyCodeRepository 验证码数据访问接口（消费者定义）.
type VerifyCodeRepository interface {
	GetVerifyCode(ctx context.Context, email string) (string, error)
	SetVerifyCode(ctx context.Context, email, code string) error
	DeleteVerifyCode(ctx context.Context, email string) error
	SetVerifyCodeRateLimit(ctx context.Context, email string) (bool, error)
}
