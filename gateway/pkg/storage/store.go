package storage

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gangantongxue/knowsync/gateway/pkg/config"
)

// Bucket 存储桶类型
type Bucket string

const (
	BucketPublic Bucket = "public" // 公共读桶，无需鉴权
	BucketAuth   Bucket = "auth"   // 认证读桶，需 JWT 临时 URL
)

// FileInfo 文件元信息
type FileInfo struct {
	Key     string    `json:"key"`      // 文件相对路径
	Size    int64     `json:"size"`     // 文件大小
	ModTime time.Time `json:"mod_time"` // 最后修改时间
	Bucket  Bucket    `json:"bucket"`   // 所属桶
}

// Store 本地文件存储，管理文件的增删查改及 JWT 临时访问
type Store struct {
	cfg       *config.Config
	jwtSecret []byte // 每次启动随机生成，重启后旧临时 URL 自动失效
}

// NewStore 创建 Store 实例
// 自动在 rootDir 下创建 public/ 和 auth/ 子目录
func NewStore(cfg *config.Config) (*Store, error) {
	jwtSecret := make([]byte, 32)
	if _, err := rand.Read(jwtSecret); err != nil {
		return nil, fmt.Errorf("生成 JWT 密钥失败: %w", err)
	}

	s := &Store{cfg: cfg, jwtSecret: jwtSecret}
	for _, b := range []Bucket{BucketPublic, BucketAuth} {
		dir := filepath.Join(cfg.Storage.RootDir, string(b))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建 %s 目录失败: %w", b, err)
		}
	}
	return s, nil
}

// ResolvePath 校验并拼接完整文件路径，防止路径穿越攻击
func (s *Store) ResolvePath(bucket Bucket, key string) (string, error) {
	bucketDir := filepath.Join(s.cfg.Storage.RootDir, string(bucket))
	cleanBucket := filepath.Clean(bucketDir)
	fullPath := filepath.Join(cleanBucket, key)

	if !strings.HasPrefix(fullPath, cleanBucket+string(filepath.Separator)) && fullPath != cleanBucket {
		return "", ErrInvalidPath
	}
	return fullPath, nil
}

const defaultJWTTTL = 30 * time.Minute

func (s *Store) jwtTTL() time.Duration {
	if s.cfg.Storage.JWTTTL <= 0 {
		return defaultJWTTTL
	}
	return s.cfg.Storage.JWTTTL
}
