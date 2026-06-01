package model

type AuthCfg struct {
	RSAPublicKeyPath string `yaml:"rsa_public_key_path" mapstructure:"rsa_public_key_path"`
}
