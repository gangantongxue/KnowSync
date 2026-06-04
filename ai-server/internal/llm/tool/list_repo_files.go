package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ListRepoFiles 查看仓库文件列表工具，实现 Eino InvokableTool 接口
type ListRepoFiles struct {
	repoDetailClient RepoDetailClient
	fileClient       FileClient
}

// NewListRepoFiles 创建查看仓库文件列表工具
func NewListRepoFiles(rdc RepoDetailClient, fc FileClient) *ListRepoFiles {
	return &ListRepoFiles{
		repoDetailClient: rdc,
		fileClient:       fc,
	}
}

// Info 返回工具元信息
func (l *ListRepoFiles) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "list_repo_files",
		Desc: "查看指定知识库中的文件目录结构。使用 repo_id 指定知识库，可选的 path 参数指定子目录路径（不传则列出根目录）。返回文件和子目录列表。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
			"path": {
				Type:     "string",
				Desc:     "目录路径，不传则列出根目录（例如：docs/）",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (l *ListRepoFiles) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return l.execute(ctx, arguments)
}

func (l *ListRepoFiles) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID string `json:"repo_id"`
		Path   string `json:"path"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{"files": [], "message": "repo_id 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"files": [], "message": "无法获取用户信息"}`, nil
	}

	// 获取仓库详情（含 owner_id，同时验证权限）
	repo, err := l.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"files": [], "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	entries, err := l.fileClient.ListRepoFiles(ctx, repo.OwnerID, params.RepoID, params.Path)
	if err != nil {
		return fmt.Sprintf(`{"files": [], "message": "获取文件列表失败: %s"}`, err.Error()), nil
	}

	type fileItem struct {
		Name string `json:"name"`
		Type string `json:"type"`
		Size int64  `json:"size"`
		Path string `json:"path"`
	}
	items := make([]fileItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, fileItem{
			Name: e.Name,
			Type: e.Type,
			Size: e.Size,
			Path: e.Path,
		})
	}

	data, _ := json.Marshal(map[string]any{
		"repo_id":   params.RepoID,
		"repo_name": repo.Name,
		"path":      params.Path,
		"files":     items,
		"total":     len(items),
	})
	return string(data), nil
}
