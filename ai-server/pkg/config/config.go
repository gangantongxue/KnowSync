package config

import (
	"log/slog"

	"github.com/spf13/viper"

	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

// NewConfig 加载并解析配置文件
// 优先级：配置文件 < 环境变量（前缀 KNOWSYNC_AI_SERVER_）
func NewConfig() (*model.Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("/app/configs")

	v.SetEnvPrefix("KNOWSYNC_AI_SERVER")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("读取配置文件失败", "error", err)
			return nil, err
		}
		slog.Warn("未找到配置文件，使用默认值和环境变量")
	}

	var cfg model.Config
	if err := v.Unmarshal(&cfg); err != nil {
		slog.Error("解析配置失败", "error", err)
		return nil, err
	}

	return &cfg, nil
}
