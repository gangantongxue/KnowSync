package storage

import (
	"os"
	"path/filepath"
	"sort"
)

// List 列出指定桶下的所有文件
// prefix 可选，不为空时只列出该前缀路径下的文件
func (s *Store) List(bucket Bucket, prefix ...string) ([]*FileInfo, error) {
	root := filepath.Join(s.cfg.Storage.RootDir, string(bucket))
	searchRoot := root
	if len(prefix) > 0 && prefix[0] != "" {
		searchRoot = filepath.Join(root, prefix[0])
	}

	var files []*FileInfo
	err := filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, &FileInfo{
			Key:     rel,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Bucket:  bucket,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Key < files[j].Key
	})
	return files, nil
}
