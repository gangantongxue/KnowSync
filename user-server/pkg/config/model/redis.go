package model

// RedisCfg Redis 配置项.
type RedisCfg struct {
	Addrs        []string `yaml:"addrs" mapstructure:"addrs"`
	Password     string   `yaml:"password" mapstructure:"password"`
	DB           int      `yaml:"db" mapstructure:"db"`
	PoolSize     int      `yaml:"pool_size" mapstructure:"pool_size"`
	MinIdleConns int      `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
}
