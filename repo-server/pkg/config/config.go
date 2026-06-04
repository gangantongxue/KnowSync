package config

import (
	"strings"

	"github.com/gangantongxue/knowsync/repo-server/pkg/config/model"
	"github.com/spf13/viper"
)

// Config repo-server 全局配置
type Config struct {
	Project  model.ProjectCfg  `yaml:"project" mapstructure:"project"`
	Logger   model.LoggerCfg   `yaml:"logger" mapstructure:"logger"`
	Database model.DatabaseCfg `yaml:"database" mapstructure:"database"`
	GRPC     model.GRPCCfg     `yaml:"grpc" mapstructure:"grpc"`
}

// NewConfig 加载配置文件并返回解析后的配置
func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("./configs")
	v.SetConfigType("yaml")

	v.SetEnvPrefix("KNOWSYNC_REPO_SERVER")
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
