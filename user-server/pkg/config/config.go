package config

import (
	"strings"

	"github.com/gangantongxue/knowsync/user-server/pkg/config/model"
	"github.com/spf13/viper"
)

// Config 配置项
type Config struct {
	Project  model.ProjectCfg  `yaml:"project"`
	Logger   model.LoggerCfg   `yaml:"logger"`
	Database model.DatabaseCfg `yaml:"database"`
	Redis    model.RedisCfg    `yaml:"redis"`
	Email    model.EmailCfg    `yaml:"email"`
	Auth     model.AuthCfg     `yaml:"auth"`
	GRPC     model.GRPCCfg     `yaml:"grpc"`
}

// NewConfig 创建一个新的配置项
func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("./configs")
	v.SetConfigType("yaml")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("KNOWSYNC_USER_SERVER")

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
		Email: model.EmailCfg{
			SMTPHost:    v.GetString("email.smtp_host"),
			SMTPPort:    v.GetInt("email.smtp_port"),
			Username:    v.GetString("email.username"),
			Password:    v.GetString("email.password"),
			FromAddress: v.GetString("email.from_address"),
			FromName:    v.GetString("email.from_name"),
		},
		Auth: model.AuthCfg{
			RSAPrivateKeyPath: v.GetString("auth.rsa_private_key_path"),
			AccessTTL:         v.GetDuration("auth.access_ttl"),
			RefreshTTL:        v.GetDuration("auth.refresh_ttl"),
		},
		GRPC: model.GRPCCfg{
			Port: v.GetInt("grpc.port"),
		},
	}

	return cfg, nil
}
