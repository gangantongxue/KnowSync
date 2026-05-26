package errcode

const (
	Success      = 200 // 成功
	ErrBadReq    = 400 // 请求参数错误
	ErrUnauth    = 401 // 未授权，令牌无效或已过期
	ErrNotFound  = 404 // 资源不存在
	ErrRateLimit = 429 // 请求频率过快
)
