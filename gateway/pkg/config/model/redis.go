package model

// RedisCfg Redis 配置
type RedisCfg struct {
	Addrs    []string `yaml:"addrs" mapstructure:"addrs"`
	Password string   `yaml:"password" mapstructure:"password"`
	DB       int      `yaml:"db" mapstructure:"db"`
}
