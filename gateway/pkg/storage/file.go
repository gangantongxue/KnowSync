package storage

import (
	"io"
	"os"
	"path/filepath"
	"time"
)

// Put 上传文件，同名覆盖
func (s *Store) Put(bucket Bucket, key string, reader io.Reader) (*FileInfo, error) {
	fullPath, err := s.ResolvePath(bucket, key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	written, err := io.Copy(f, reader)
	if err != nil {
		return nil, err
	}
	return &FileInfo{Key: key, Size: written, ModTime: time.Now(), Bucket: bucket}, nil
}

// PutIfAbsent 上传文件，文件已存在时返回 ErrFileAlreadyExists
// 使用 O_CREAT|O_EXCL 原子操作创建，避免 TOCTOU 竞态
func (s *Store) PutIfAbsent(bucket Bucket, key string, reader io.Reader) (*FileInfo, error) {
	fullPath, err := s.ResolvePath(bucket, key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrFileAlreadyExists
		}
		return nil, err
	}
	defer f.Close()

	written, err := io.Copy(f, reader)
	if err != nil {
		return nil, err
	}
	return &FileInfo{Key: key, Size: written, ModTime: time.Now(), Bucket: bucket}, nil
}

// Get 读取文件内容
func (s *Store) Get(bucket Bucket, key string) (io.ReadCloser, error) {
	fullPath, err := s.ResolvePath(bucket, key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return f, nil
}

// Delete 删除单个文件
func (s *Store) Delete(bucket Bucket, key string) error {
	fullPath, err := s.ResolvePath(bucket, key)
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

// DeleteDir 递归删除目录及其所有子文件和子目录
func (s *Store) DeleteDir(bucket Bucket, dirPath string) error {
	bucketDir := filepath.Join(s.cfg.Storage.RootDir, string(bucket))
	cleanBucket := filepath.Clean(bucketDir)
	target := filepath.Join(cleanBucket, dirPath)

	if target == cleanBucket {
		return ErrInvalidPath
	}

	return os.RemoveAll(target)
}
