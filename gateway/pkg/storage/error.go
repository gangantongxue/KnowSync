package storage

import "errors"

var (
	ErrFileNotFound       = errors.New("文件不存在")
	ErrFileAlreadyExists  = errors.New("文件已存在")
	ErrInvalidPath        = errors.New("非法的文件路径")
	ErrInvalidToken       = errors.New("无效或已过期的令牌")
	ErrPublicBucketNoSign = errors.New("公共读桶无需签发令牌")
)
