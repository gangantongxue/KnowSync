package config

import (
	"strings"

	"github.com/gangantongxue/knowsync/chat-server/pkg/config/model"
	"github.com/spf13/viper"
)

// Config 配置项
type Config struct {
	Project  model.ProjectCfg  `yaml:"project" mapstructure:"project"`
	Logger   model.LoggerCfg   `yaml:"logger" mapstructure:"logger"`
	Database model.DatabaseCfg `yaml:"database" mapstructure:"database"`
	Redis    model.RedisCfg    `yaml:"redis" mapstructure:"redis"`
	GRPC     model.GRPCCfg     `yaml:"grpc" mapstructure:"grpc"`
	WS       model.WSCfg       `yaml:"ws" mapstructure:"ws"`
	JWT      model.JWTCfg      `yaml:"jwt" mapstructure:"jwt"`
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

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
