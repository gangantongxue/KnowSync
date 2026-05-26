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
}

// NewConfig 创建一个新的配置项
func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("./configs")
	v.SetConfigType("yaml")

	v.SetEnvPrefix("KNOWSYNC_USER_SERVER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
