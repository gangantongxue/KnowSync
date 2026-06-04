package tool //nolint:dupl // 与 add_collaborator.go 结构相似但逻辑不同

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// UpdateCollaboratorRole 变更协作者角色工具.
type UpdateCollaboratorRole struct {
	repoDetailClient   RepoDetailClient
	collaboratorClient CollaboratorClient
}

// NewUpdateCollaboratorRole 创建 UpdateCollaboratorRole 工具.
func NewUpdateCollaboratorRole(rdc RepoDetailClient, cc CollaboratorClient) *UpdateCollaboratorRole {
	return &UpdateCollaboratorRole{
		repoDetailClient:   rdc,
		collaboratorClient: cc,
	}
}

// Info 返回工具元信息.
func (u *UpdateCollaboratorRole) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_collaborator_role",
		Desc: "变更知识库协作者的角色。务必先用 ask_user 让用户确认后再执行。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamRepoID: {
				Type:     TypeString,
				Desc:     DescRepoID,
				Required: true,
			},
			ParamUserID: {
				Type:     TypeString,
				Desc:     "协作者用户 ID",
				Required: true,
			},
			ParamRole: {
				Type:     TypeString,
				Desc:     "新角色：ADMIN、DEVELOPER 或 VIEWER",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (u *UpdateCollaboratorRole) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
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
		return ErrRespRepoIDEmpty, nil
	}
	if params.UserID == "" {
		return `{"success": false, "message": "user_id 不能为空"}`, nil
	}
	if params.Role == "" {
		return `{"success": false, "message": "role 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return ErrRespUserInfo, nil
	}

	// 验证仓库存在且有权限
	if _, err := u.repoDetailClient.GetRepo(ctx, params.RepoID, userID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := u.collaboratorClient.UpdateCollaboratorRole(ctx, params.RepoID, params.UserID, params.Role); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "变更协作者角色失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		KeySuccess:  true,
		ParamRepoID: params.RepoID,
		ParamUserID: params.UserID,
		ParamRole:   params.Role,
	})
	return string(data), nil
}
