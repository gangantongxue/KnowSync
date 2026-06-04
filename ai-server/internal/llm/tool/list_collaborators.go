package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ListCollaborators 查看协作者列表工具.
type ListCollaborators struct {
	repoDetailClient   RepoDetailClient
	collaboratorClient CollaboratorClient
}

// NewListCollaborators 创建 ListCollaborators 工具.
func NewListCollaborators(rdc RepoDetailClient, cc CollaboratorClient) *ListCollaborators {
	return &ListCollaborators{
		repoDetailClient:   rdc,
		collaboratorClient: cc,
	}
}

// Info 返回工具元信息.
func (l *ListCollaborators) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "list_collaborators",
		Desc: "查看知识库的协作者列表。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamRepoID: {
				Type:     TypeString,
				Desc:     DescRepoID,
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (l *ListCollaborators) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return l.execute(ctx, arguments)
}

func (l *ListCollaborators) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID string `json:"repo_id"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{"collaborators": [], "message": "repo_id 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"collaborators": [], "message": "无法获取用户信息"}`, nil
	}

	// 验证仓库存在且有权限
	repo, err := l.repoDetailClient.GetRepo(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"collaborators": [], "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	collaborators, err := l.collaboratorClient.ListCollaborators(ctx, params.RepoID)
	if err != nil {
		return fmt.Sprintf(`{"collaborators": [], "message": "获取协作者列表失败: %s"}`, err.Error()), nil
	}

	type collabItem struct {
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
		Role     string `json:"role"`
	}
	items := make([]collabItem, 0, len(collaborators))
	for _, c := range collaborators {
		items = append(items, collabItem(c))
	}

	data, _ := json.Marshal(map[string]any{
		ParamRepoID:     params.RepoID,
		"repo_name":     repo.Name,
		"collaborators": items,
		KeyTotal:        len(items),
	})
	return string(data), nil
}
