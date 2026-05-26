package redis

import (
	"context"
	"errors"
	"time"

	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisClient 定义通用接口
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

// Redis Redis 连接
type Redis struct {
	RDB    RedisClient
	Cfg    *config.Config
	Logger *logger.Logger
}

// NewRedis 创建一个新的 Redis 连接
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
