package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// UpdateCollaboratorRole 变更协作者角色工具
type UpdateCollaboratorRole struct {
	repoDetailClient   RepoDetailClient
	collaboratorClient CollaboratorClient
}

// NewUpdateCollaboratorRole 创建 UpdateCollaboratorRole 工具
func NewUpdateCollaboratorRole(rdc RepoDetailClient, cc CollaboratorClient) *UpdateCollaboratorRole {
	return &UpdateCollaboratorRole{
		repoDetailClient:   rdc,
		collaboratorClient: cc,
	}
}

// Info 返回工具元信息
func (u *UpdateCollaboratorRole) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_collaborator_role",
		Desc: "变更知识库协作者的角色。务必先用 ask_user 让用户确认后再执行。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
			"user_id": {
				Type:     "string",
				Desc:     "协作者用户 ID",
				Required: true,
			},
			"role": {
				Type:     "string",
				Desc:     "新角色：ADMIN、DEVELOPER 或 VIEWER",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (u *UpdateCollaboratorRole) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return u.execute(ctx, arguments)
}

func (u *UpdateCollaboratorRole) execute(ctx context.Context, paramsJSON string) (string, error) {
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
	if _, err := u.repoDetailClient.GetRepo(ctx, params.RepoID, userID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := u.collaboratorClient.UpdateCollaboratorRole(ctx, params.RepoID, params.UserID, params.Role); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "变更协作者角色失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"success": true,
		"repo_id": params.RepoID,
		"user_id": params.UserID,
		"role":    params.Role,
	})
	return string(data), nil
}
