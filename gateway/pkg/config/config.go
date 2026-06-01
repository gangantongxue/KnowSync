package config

import (
	"strings"

	"github.com/gangantongxue/knowsync/gateway/pkg/config/model"
	"github.com/spf13/viper"
)

// Config 网关全局配置
type Config struct {
	Project     model.ProjectCfg     `yaml:"project" mapstructure:"project"`
	Logger      model.LoggerCfg      `yaml:"logger" mapstructure:"logger"`
	HTTP        model.HTTPCfg        `yaml:"http" mapstructure:"http"`
	Auth        model.AuthCfg        `yaml:"auth" mapstructure:"auth"`
	GRPCClients model.GRPCClientsCfg `yaml:"grpc_clients" mapstructure:"grpc_clients"`
	Storage     model.StorageCfg     `yaml:"storage" mapstructure:"storage"`
}

// NewConfig 读取 configs/config.yaml 并返回解析后的配置
func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("./configs")
	v.SetConfigType("yaml")

	v.SetEnvPrefix("KNOWSYNC_GATEWAY")
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
