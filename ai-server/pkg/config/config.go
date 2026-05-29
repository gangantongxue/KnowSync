package config

import (
	"log/slog"
	"strings"

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

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("KNOWSYNC_AI_SERVER")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("读取配置文件失败", "error", err)
			return nil, err
		}
		slog.Warn("未找到配置文件，使用默认值和环境变量")
	}

	cfg := &model.Config{
		GRPC: model.GRPCCfg{
			Port: v.GetInt("grpc.port"),
		},
		Log: model.LogCfg{
			Level: v.GetString("log.level"),
			Dir:   v.GetString("log.dir"),
		},
		Redis: model.RedisCfg{
			Addrs:    v.GetStringSlice("redis.addrs"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
		Database: model.DatabaseCfg{
			Host:     v.GetString("database.host"),
			Port:     v.GetInt("database.port"),
			User:     v.GetString("database.user"),
			Password: v.GetString("database.password"),
			DBName:   v.GetString("database.dbname"),
		},
		RepoServer: model.RepoServerCfg{
			Addr: v.GetString("repo_server.addr"),
		},
		Gateway: model.GatewayCfg{
			Addr: v.GetString("gateway.addr"),
		},
		Embedder: model.EmbedderCfg{
			BaseURL:    v.GetString("embedder.base_url"),
			APIKey:     v.GetString("embedder.api_key"),
			Model:      v.GetString("embedder.model"),
			Dimensions: v.GetInt("embedder.dimensions"),
		},
		LLM: model.LLMCfg{
			BaseURL:           v.GetString("llm.base_url"),
			APIKey:            v.GetString("llm.api_key"),
			Model:             v.GetString("llm.model"),
			ThinkingIntensity: v.GetString("llm.thinking_intensity"),
		},
		Chromem: model.ChromemCfg{
			Path: v.GetString("chromem.path"),
		},
		Worker: model.WorkerCfg{
			Concurrency: v.GetInt("worker.concurrency"),
			MaxRetries:  v.GetInt("worker.max_retries"),
		},
	}

	return cfg, nil
}
