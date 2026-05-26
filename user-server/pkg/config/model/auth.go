package model

import "time"

// AuthCfg 认证配置
type AuthCfg struct {
	JWTSecret  string        `yaml:"jwt_secret"`
	AccessTTL  time.Duration `yaml:"access_ttl"`
	RefreshTTL time.Duration `yaml:"refresh_ttl"`
}
