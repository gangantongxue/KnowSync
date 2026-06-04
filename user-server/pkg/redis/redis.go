// Package redis provides Redis client connections and operations.
package redis

import (
	"context"
	"errors"
	"time"

	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Client 定义通用 Redis 客户端接口.
type Client interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

// Redis 连接.
type Redis struct {
	RDB    Client
	Cfg    *config.Config
	Logger *logger.Logger
}

// NewRedis 创建一个新的 Redis 连接.
func NewRedis(cfg *config.Config, logger *logger.Logger) (*Redis, error) {
	if len(cfg.Redis.Addrs) == 0 {
		return nil, errors.New("redis addrs is empty")
	}

	r := &Redis{
		Cfg:    cfg,
		Logger: logger,
	}

	// 设置 Redis 日志适配器
	redis.SetLogger(logger)

	// 判断 Redis 单点或集群
	if len(cfg.Redis.Addrs) == 1 {
		// 单点 Redis
		rdb := redis.NewClient(&redis.Options{
			Addr:         cfg.Redis.Addrs[0],
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
			PoolSize:     cfg.Redis.PoolSize,
			MinIdleConns: cfg.Redis.MinIdleConns,
		})
		r.RDB = rdb
	} else {
		// 集群 Redis
		rdb := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Redis.Addrs,
			Password:     cfg.Redis.Password,
			PoolSize:     cfg.Redis.PoolSize,
			MinIdleConns: cfg.Redis.MinIdleConns,
		})
		r.RDB = rdb
	}

	return r, nil
}
