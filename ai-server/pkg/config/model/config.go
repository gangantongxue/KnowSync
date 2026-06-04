package model

// Config 应用配置
type Config struct {
	GRPC         GRPCCfg         `yaml:"grpc"`
	Log          LogCfg          `yaml:"log"`
	Redis        RedisCfg        `yaml:"redis"`
	Database     DatabaseCfg     `yaml:"database"`
	Gateway      GatewayCfg      `yaml:"gateway"`
	Embedder     EmbedderCfg     `yaml:"embedder"`
	LLM          LLMCfg          `yaml:"llm"`
	Chromem      ChromemCfg      `yaml:"chromem"`
	Worker       WorkerCfg       `yaml:"worker"`
	ServiceToken ServiceTokenCfg `yaml:"service_token"`
}

// GRPCCfg gRPC 服务器配置
type GRPCCfg struct {
	Port int `yaml:"port"`
}

// LogCfg 日志配置
type LogCfg struct {
	Level string `yaml:"level"`
	Dir   string `yaml:"dir"`
}

// RedisCfg Redis 配置
type RedisCfg struct {
	Addrs    []string `yaml:"addrs"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
}

// DatabaseCfg 数据库配置
type DatabaseCfg struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// ServiceTokenCfg service token 配置
type ServiceTokenCfg struct {
	Secret string `yaml:"secret"`
}

// GatewayCfg gateway HTTP 客户端配置
type GatewayCfg struct {
	Addr string `yaml:"addr"`
}

// EmbedderCfg 向量化模型配置
type EmbedderCfg struct {
	BaseURL    string `yaml:"base_url"`
	APIKey     string `yaml:"api_key"`
	Model      string `yaml:"model"`
	Dimensions int    `yaml:"dimensions"`
}

// LLMCfg 大语言模型配置
type LLMCfg struct {
	BaseURL           string `yaml:"base_url"`
	APIKey            string `yaml:"api_key"`
	Model             string `yaml:"model"`
	ThinkingIntensity string `yaml:"thinking_intensity"`
}

// ChromemCfg chromem-go 向量存储配置
type ChromemCfg struct {
	Path string `yaml:"path"`
}

// WorkerCfg 后台 Worker 配置
type WorkerCfg struct {
	Concurrency int `yaml:"concurrency"`
	MaxRetries  int `yaml:"max_retries"`
}
