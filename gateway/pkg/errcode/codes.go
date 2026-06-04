// Package errcode defines business error codes for the gateway.
package errcode

const (
	// Success 表示请求处理成功.
	Success = 200 // 成功
	// ErrBadReq 表示请求参数错误.
	ErrBadReq = 400 // 请求参数错误
	// ErrUnauth 表示未授权，令牌无效或已过期.
	ErrUnauth = 401 // 未授权，令牌无效或已过期
	// ErrNotFound 表示资源不存在.
	ErrNotFound = 404 // 资源不存在
	// ErrRateLimit 表示请求频率过快.
	ErrRateLimit = 429 // 请求频率过快
)
