// Package model provides configuration data structures.
package model

import "time"

// AuthCfg 认证配置.
type AuthCfg struct {
	RSAPrivateKeyPath string        `yaml:"rsa_private_key_path" mapstructure:"rsa_private_key_path"`
	AccessTTL         time.Duration `yaml:"access_ttl" mapstructure:"access_ttl"`
	RefreshTTL        time.Duration `yaml:"refresh_ttl" mapstructure:"refresh_ttl"`
}
