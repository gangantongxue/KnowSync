package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// GetFileContent 获取仓库文件内容工具，实现 Eino InvokableTool 接口.
type GetFileContent struct {
	repoDetailClient RepoDetailClient
	fileClient       FileClient
}

// NewGetFileContent 创建获取仓库文件内容工具.
func NewGetFileContent(rdc RepoDetailClient, fc FileClient) *GetFileContent {
	return &GetFileContent{
		repoDetailClient: rdc,
		fileClient:       fc,
	}
}

// Info 返回工具元信息.
func (g *GetFileContent) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_file_content",
		Desc: "读取知识库中指定文件的完整内容。使用 repo_id 指定知识库，file_path 指定文件路径（从 list_repo_files 获取得的 path）。适用于需要深入了解某个文件的具体内容时。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamRepoID: {
				Type:     TypeString,
				Desc:     DescRepoID,
				Required: true,
			},
			ParamFilePath: {
				Type:     TypeString,
				Desc:     DescFilePath,
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (g *GetFileContent) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return g.execute(ctx, arguments)
}

func (g *GetFileContent) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID   string `json:"repo_id"`
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{ParamContent: "", "message": "repo_id 不能为空"}`, nil
	}
	if params.FilePath == "" {
		return `{ParamContent: "", "message": "file_path 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{ParamContent: "", "message": "无法获取用户信息"}`, nil
	}

	// 获取仓库详情（含 owner_id，同时验证权限）
	repo, err := g.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{ParamContent: "", "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	content, err := g.fileClient.GetFileContent(ctx, repo.OwnerID, params.RepoID, params.FilePath)
	if err != nil {
		return fmt.Sprintf(`{ParamContent: "", "message": "读取文件内容失败: %s"}`, err.Error()), nil
	}

	if content == "" {
		return `{ParamContent: "", "message": "文件内容为空"}`, nil
	}

	data, _ := json.Marshal(map[string]any{
		ParamRepoID:   params.RepoID,
		ParamFilePath: params.FilePath,
		ParamContent:  content,
		ParamSize:     len(content),
	})
	return string(data), nil
}
