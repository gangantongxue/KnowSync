// Package storage provides local file storage operations.
package storage

import "errors"

var (
	// ErrFileNotFound 表示文件不存在的错误.
	ErrFileNotFound = errors.New("文件不存在")
	// ErrFileAlreadyExists 表示文件已存在的错误.
	ErrFileAlreadyExists = errors.New("文件已存在")
	// ErrInvalidPath 表示非法的文件路径错误.
	ErrInvalidPath = errors.New("非法的文件路径")
)
