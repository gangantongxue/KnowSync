package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// AddCollaborator 添加协作者工具
type AddCollaborator struct {
	repoDetailClient   RepoDetailClient
	collaboratorClient CollaboratorClient
}

// NewAddCollaborator 创建 AddCollaborator 工具
func NewAddCollaborator(rdc RepoDetailClient, cc CollaboratorClient) *AddCollaborator {
	return &AddCollaborator{
		repoDetailClient:   rdc,
		collaboratorClient: cc,
	}
}

// Info 返回工具元信息
func (a *AddCollaborator) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "add_collaborator",
		Desc: "向知识库添加协作者。先通过 search_users 找到用户 ID，再用此工具添加。务必先用 ask_user 让用户确认后再执行。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
			"user_id": {
				Type:     "string",
				Desc:     "用户 ID（通过 search_users 获取）",
				Required: true,
			},
			"role": {
				Type:     "string",
				Desc:     "角色：ADMIN、DEVELOPER 或 VIEWER",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (a *AddCollaborator) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return a.execute(ctx, arguments)
}

func (a *AddCollaborator) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID string `json:"repo_id"`
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{"success": false, "message": "repo_id 不能为空"}`, nil
	}
	if params.UserID == "" {
		return `{"success": false, "message": "user_id 不能为空"}`, nil
	}
	if params.Role == "" {
		return `{"success": false, "message": "role 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"success": false, "message": "无法获取用户信息"}`, nil
	}

	// 验证仓库存在且有权限
	if _, err := a.repoDetailClient.GetRepo(ctx, params.RepoID, userID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := a.collaboratorClient.AddCollaborator(ctx, params.RepoID, params.UserID, params.Role); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "添加协作者失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"success": true,
		"repo_id": params.RepoID,
		"user_id": params.UserID,
		"role":    params.Role,
	})
	return string(data), nil
}
