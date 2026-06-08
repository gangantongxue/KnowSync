package service

import "errors"

var (
	// ErrPermissionDenied 权限不足错误.
	ErrPermissionDenied = errors.New("权限不足")
	// ErrNotFound 资源不存在错误.
	ErrNotFound = errors.New("资源不存在")
	// ErrAlreadyExists 资源已存在错误.
	ErrAlreadyExists = errors.New("已存在同名资源")
)
