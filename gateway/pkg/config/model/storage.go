package model

// StorageCfg 本地文件存储配置.
type StorageCfg struct {
	RootDir string `yaml:"root_dir" mapstructure:"root_dir"` // 文件存储根目录
}
