package model

import "time"

// StorageCfg 本地文件存储配置
type StorageCfg struct {
	RootDir string        `yaml:"root_dir"` // 文件存储根目录
	JWTTTL  time.Duration `yaml:"jwt_ttl"`  // 临时访问 URL 有效期
}
