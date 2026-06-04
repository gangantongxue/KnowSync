// Package redis 提供 Redis 客户端初始化和管理.
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

// NewRedis 创建 Redis 客户端.
func NewRedis(cfg *model.RedisCfg) (*redis.Client, error) {
	addr := "localhost:6379"
	if len(cfg.Addrs) > 0 {
		addr = cfg.Addrs[0]
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return rdb, nil
}
