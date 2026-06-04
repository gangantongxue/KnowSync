// Package model 提供 AI 服务配置结构体定义.
package model

// Config 应用配置.
type Config struct {
	GRPC         GRPCCfg         `yaml:"grpc" mapstructure:"grpc"`
	Log          LogCfg          `yaml:"log" mapstructure:"log"`
	Redis        RedisCfg        `yaml:"redis" mapstructure:"redis"`
	Database     DatabaseCfg     `yaml:"database" mapstructure:"database"`
	Gateway      GatewayCfg      `yaml:"gateway" mapstructure:"gateway"`
	Embedder     EmbedderCfg     `yaml:"embedder" mapstructure:"embedder"`
	LLM          LLMCfg          `yaml:"llm" mapstructure:"llm"`
	Chromem      ChromemCfg      `yaml:"chromem" mapstructure:"chromem"`
	Worker       WorkerCfg       `yaml:"worker" mapstructure:"worker"`
	ServiceToken ServiceTokenCfg `yaml:"service_token" mapstructure:"service_token"`
	WebSearch    WebSearchCfg    `yaml:"web_search" mapstructure:"web_search"`
}

// GRPCCfg gRPC 服务器配置.
type GRPCCfg struct {
	Port int `yaml:"port" mapstructure:"port"`
}

// LogCfg 日志配置.
type LogCfg struct {
	Level string `yaml:"level" mapstructure:"level"`
	Dir   string `yaml:"dir" mapstructure:"dir"`
}

// RedisCfg Redis 配置.
type RedisCfg struct {
	Addrs    []string `yaml:"addrs" mapstructure:"addrs"`
	Password string   `yaml:"password" mapstructure:"password"`
	DB       int      `yaml:"db" mapstructure:"db"`
}

// DatabaseCfg 数据库配置.
type DatabaseCfg struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	DBName   string `yaml:"dbname" mapstructure:"dbname"`
}

// ServiceTokenCfg service token 配置.
type ServiceTokenCfg struct {
	Secret string `yaml:"secret" mapstructure:"secret"`
}

// GatewayCfg gateway HTTP 客户端配置.
type GatewayCfg struct {
	Addr string `yaml:"addr" mapstructure:"addr"`
}

// EmbedderCfg 向量化模型配置.
type EmbedderCfg struct {
	BaseURL    string `yaml:"base_url" mapstructure:"base_url"`
	APIKey     string `yaml:"api_key" mapstructure:"api_key"`
	Model      string `yaml:"model" mapstructure:"model"`
	Dimensions int    `yaml:"dimensions" mapstructure:"dimensions"`
}

// LLMCfg 大语言模型配置.
type LLMCfg struct {
	BaseURL           string `yaml:"base_url" mapstructure:"base_url"`
	APIKey            string `yaml:"api_key" mapstructure:"api_key"`
	Model             string `yaml:"model" mapstructure:"model"`
	ThinkingIntensity string `yaml:"thinking_intensity" mapstructure:"thinking_intensity"`
}

// ChromemCfg chromem-go 向量存储配置.
type ChromemCfg struct {
	Path string `yaml:"path" mapstructure:"path"`
}

// WorkerCfg 后台 Worker 配置.
type WorkerCfg struct {
	Concurrency int `yaml:"concurrency" mapstructure:"concurrency"`
	MaxRetries  int `yaml:"max_retries" mapstructure:"max_retries"`
}

// WebSearchCfg 网络搜索配置.
type WebSearchCfg struct {
	BaseURL string `yaml:"base_url" mapstructure:"base_url"` // 搜索 API 地址（默认 https://api.duckduckgo.com/）
	APIKey  string `yaml:"api_key" mapstructure:"api_key"`   // 搜索 API Key（可选，部分搜索服务需要）
}
