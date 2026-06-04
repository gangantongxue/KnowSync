package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// RemoveCollaborator 移除协作者工具
type RemoveCollaborator struct {
	repoDetailClient   RepoDetailClient
	collaboratorClient CollaboratorClient
}

// NewRemoveCollaborator 创建 RemoveCollaborator 工具
func NewRemoveCollaborator(rdc RepoDetailClient, cc CollaboratorClient) *RemoveCollaborator {
	return &RemoveCollaborator{
		repoDetailClient:   rdc,
		collaboratorClient: cc,
	}
}

// Info 返回工具元信息
func (r *RemoveCollaborator) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "remove_collaborator",
		Desc: "移除知识库的协作者。务必先用 ask_user 让用户确认后再执行。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
			"user_id": {
				Type:     "string",
				Desc:     "要移除的协作者用户 ID",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (r *RemoveCollaborator) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return r.execute(ctx, arguments)
}

func (r *RemoveCollaborator) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID string `json:"repo_id"`
		UserID string `json:"user_id"`
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

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"success": false, "message": "无法获取用户信息"}`, nil
	}

	// 验证仓库存在且有权限
	if _, err := r.repoDetailClient.GetRepo(ctx, params.RepoID, userID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := r.collaboratorClient.RemoveCollaborator(ctx, params.RepoID, params.UserID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "移除协作者失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"success": true,
		"repo_id": params.RepoID,
		"user_id": params.UserID,
	})
	return string(data), nil
}
