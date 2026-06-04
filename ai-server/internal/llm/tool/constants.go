// Package tool common constants for tool parameter names and error messages.
package tool

// 常用 JSON 键名.
const (
	KeySuccess = "success"
	KeyMessage = "message"
	KeyTotal   = "total"
	KeyFound   = "found"
	KeyRepos   = "repos"
	KeyData    = "data"
)

// 常用参数名.
const (
	ParamRepoID   = "repo_id"
	ParamUserID   = "user_id"
	ParamFilePath = "file_path"
	ParamRole     = "role"
	ParamName     = "name"
	ParamContent  = "content"
	ParamDesc     = "description"
	ParamSkipCfm  = "_skip_confirm"
	ParamType     = "type"
	ParamSize     = "size"
	ParamVisibl   = "visibility"
)

// 常用参数类型.
const (
	TypeString  = "string"
	TypeBoolean = "boolean"
)

// 常用错误信息.
const (
	ErrRepoIDEmpty  = "repo_id 不能为空"
	ErrUserIDEmpty  = "user_id 不能为空"
	ErrFilePathE    = "file_path 不能为空"
	ErrContentEmpty = "content 不能为空"
	ErrRoleEmpty    = "role 不能为空"
	ErrNameEmpty    = "name 不能为空"
	ErrURLRequired  = "url 不能为空"
	ErrTypeRequired = "type 不能为空"
	ErrUserInfo     = "无法获取用户信息"
)

// 参数描述.
const (
	DescFileContent = "文件内容（Markdown 格式）"
	DescNoDesc      = "暂无描述"
	DescRepoID      = "知识库 ID"
	DescFilePath    = "文件路径，例如：docs/chapter1.md"
	DescSkipConfirm = "用户已明确确认时设置为 true"
)

// 错误响应 JSON（使用 fmt.Sprintf 注入变量，或直接用于精确匹配的错误场景）.
const (
	ErrRespRepoIDEmpty = `{"success": false, "message": "repo_id 不能为空"}`
	ErrRespUserInfo    = `{"success": false, "message": "无法获取用户信息"}`
)
