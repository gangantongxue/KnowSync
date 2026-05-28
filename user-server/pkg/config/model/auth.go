package model

import "time"

type AuthCfg struct {
	RSAPrivateKeyPath string        `yaml:"rsa_private_key_path"`
	AccessTTL         time.Duration `yaml:"access_ttl"`
	RefreshTTL        time.Duration `yaml:"refresh_ttl"`
}
