package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DirEntry 目录条目信息.
type DirEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// Store 本地文件存储，管理文件的增删查改.
type Store struct {
	root string // 存储根目录绝对路径
}

// NewStore 创建 Store 实例.
func NewStore(rootDir string) (*Store, error) {
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, err
	}
	return &Store{root: abs}, nil
}

// Resolve 校验并拼接完整文件路径，防止路径穿越攻击.
func (s *Store) Resolve(subpath string) (string, error) {
	abs := filepath.Join(s.root, subpath)
	abs = filepath.Clean(abs)
	if !strings.HasPrefix(abs, s.root+string(filepath.Separator)) && abs != s.root {
		return "", ErrInvalidPath
	}
	return abs, nil
}

// WriteFile 写入文件内容，自动创建父目录.
func (s *Store) WriteFile(subpath string, reader io.Reader) (int64, error) {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return 0, err
	}
	f, err := os.Create(fullPath) //nolint:gosec // path is sanitized by Resolve
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	return io.Copy(f, reader)
}

// WriteFileFromBytes 写入文件内容（字节数组形式），自动创建父目录.
func (s *Store) WriteFileFromBytes(subpath string, data []byte) error {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0o600)
}

// ReadFile 打开文件用于读取.
func (s *Store) ReadFile(subpath string) (io.ReadCloser, error) {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fullPath) //nolint:gosec // path is sanitized by Resolve
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return f, nil
}

// Delete 删除单个文件.
func (s *Store) Delete(subpath string) error {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}
	return nil
}

// DeleteAll 递归删除目录及其所有内容.
func (s *Store) DeleteAll(subpath string) error {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return err
	}
	if fullPath == s.root {
		return ErrInvalidPath
	}
	return os.RemoveAll(fullPath)
}

// Rename 移动/重命名文件或目录.
func (s *Store) Rename(oldSubpath, newSubpath string) error {
	oldFull, err := s.Resolve(oldSubpath)
	if err != nil {
		return err
	}
	newFull, err := s.Resolve(newSubpath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(newFull), 0o750); err != nil {
		return err
	}
	return os.Rename(oldFull, newFull)
}

// ListDir 列出目录内容，自动过滤 .hertz.gz 等临时文件.
func (s *Store) ListDir(subpath string) ([]DirEntry, error) {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	result := make([]DirEntry, 0, len(entries))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".hertz.gz") {
			continue
		}
		info, _ := e.Info()
		var size int64
		if info != nil {
			size = info.Size()
		}
		result = append(result, DirEntry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Size:  size,
		})
	}
	return result, nil
}

// Stat 获取文件信息.
func (s *Store) Stat(subpath string) (os.FileInfo, error) {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return info, nil
}

// MakeDir 创建目录（含父目录）.
func (s *Store) MakeDir(subpath string) error {
	fullPath, err := s.Resolve(subpath)
	if err != nil {
		return err
	}
	return os.MkdirAll(fullPath, 0o750)
}
