package model

// ProjectCfg 项目配置项.
type ProjectCfg struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Version string `yaml:"version" mapstructure:"version"`
}
