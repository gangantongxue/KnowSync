package model

// ServiceTokenCfg service token 配置.
type ServiceTokenCfg struct {
	Secret string `yaml:"secret" mapstructure:"secret"`
}
