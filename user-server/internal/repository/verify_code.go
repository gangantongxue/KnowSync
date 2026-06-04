package repository

import (
	"context"
	"fmt"
	"time"
)

const (
	// verifyCodePrefix 验证码 Redis key 前缀.
	verifyCodePrefix = "verify_code:"
	// verifyCodeTTL 验证码有效期.
	verifyCodeTTL = 5 * time.Minute
	// rateLimitPrefix 发送频率限制 Redis key 前缀.
	rateLimitPrefix = "rate_limit:verify_code:"
	// rateLimitTTL 发送频率限制时间窗口.
	rateLimitTTL = 60 * time.Second
)

// SetVerifyCode 将验证码存入 Redis，设置过期时间.
func (r *Repository) SetVerifyCode(ctx context.Context, email, code string) error {
	key := fmt.Sprintf("%s%s", verifyCodePrefix, email)
	return r.Redis.RDB.Set(ctx, key, code, verifyCodeTTL).Err()
}

// GetVerifyCode 从 Redis 获取验证码.
func (r *Repository) GetVerifyCode(ctx context.Context, email string) (string, error) {
	key := fmt.Sprintf("%s%s", verifyCodePrefix, email)
	return r.Redis.RDB.Get(ctx, key).Result()
}

// DeleteVerifyCode 从 Redis 删除验证码.
func (r *Repository) DeleteVerifyCode(ctx context.Context, email string) error {
	key := fmt.Sprintf("%s%s", verifyCodePrefix, email)
	return r.Redis.RDB.Del(ctx, key).Err()
}

// SetVerifyCodeRateLimit 尝试设置发送频率限制，返回是否允许发送
// 使用 SET NX + EX 原子操作，key 已存在时返回 false 表示频率过高.
func (r *Repository) SetVerifyCodeRateLimit(ctx context.Context, email string) (bool, error) {
	key := fmt.Sprintf("%s%s", rateLimitPrefix, email)
	return r.Redis.RDB.SetNX(ctx, key, "1", rateLimitTTL).Result()
}
