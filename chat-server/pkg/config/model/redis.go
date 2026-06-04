package model

// RedisCfg Redis 配置项
type RedisCfg struct {
	Addrs        []string `yaml:"addrs"`
	Password     string   `yaml:"password"`
	DB           int      `yaml:"db"`
	PoolSize     int      `yaml:"pool_size"`
	MinIdleConns int      `yaml:"min_idle_conns"`
}
