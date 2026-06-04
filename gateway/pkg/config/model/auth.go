// Package model provides configuration data structures.
package model

// AuthCfg holds RSA authentication configuration.
type AuthCfg struct {
	RSAPublicKeyPath string `yaml:"rsa_public_key_path" mapstructure:"rsa_public_key_path"`
}
