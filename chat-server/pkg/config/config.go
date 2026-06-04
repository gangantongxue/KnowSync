package config

import (
	"strings"

	"github.com/gangantongxue/knowsync/chat-server/pkg/config/model"
	"github.com/spf13/viper"
)

// Config 配置项
type Config struct {
	Project  model.ProjectCfg  `yaml:"project"`
	Logger   model.LoggerCfg   `yaml:"logger"`
	Database model.DatabaseCfg `yaml:"database"`
	Redis    model.RedisCfg    `yaml:"redis"`
	GRPC     model.GRPCCfg     `yaml:"grpc"`
	WS       model.WSCfg       `yaml:"ws"`
	JWT      model.JWTCfg      `yaml:"jwt"`
}

// NewConfig 创建一个新的配置项
func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("./configs")
	v.SetConfigType("yaml")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("KNOWSYNC_CHAT_SERVER")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{
		Project: model.ProjectCfg{
			Name:    v.GetString("project.name"),
			Version: v.GetString("project.version"),
		},
		Logger: model.LoggerCfg{
			Level:      v.GetString("logger.level"),
			Dir:        v.GetString("logger.dir"),
			MaxSize:    v.GetInt("logger.max_size"),
			MaxBackups: v.GetInt("logger.max_backups"),
			MaxAge:     v.GetInt("logger.max_age"),
			Compress:   v.GetBool("logger.compress"),
			LocalTime:  v.GetBool("logger.local_time"),
		},
		Database: model.DatabaseCfg{
			Host:            v.GetString("database.host"),
			Port:            v.GetInt("database.port"),
			User:            v.GetString("database.user"),
			Password:        v.GetString("database.password"),
			DBName:          v.GetString("database.dbname"),
			MaxIdleConns:    v.GetInt("database.max_idle_conns"),
			MaxOpenConns:    v.GetInt("database.max_open_conns"),
			ConnMaxLifetime: v.GetDuration("database.conn_max_lifetime"),
		},
		Redis: model.RedisCfg{
			Addrs:        v.GetStringSlice("redis.addrs"),
			Password:     v.GetString("redis.password"),
			DB:           v.GetInt("redis.db"),
			PoolSize:     v.GetInt("redis.pool_size"),
			MinIdleConns: v.GetInt("redis.min_idle_conns"),
		},
		GRPC: model.GRPCCfg{
			Port: v.GetInt("grpc.port"),
		},
		WS: model.WSCfg{
			Port: v.GetInt("ws.port"),
		},
		JWT: model.JWTCfg{
			PublicKeyPath: v.GetString("jwt.public_key_path"),
		},
	}

	return cfg, nil
}
