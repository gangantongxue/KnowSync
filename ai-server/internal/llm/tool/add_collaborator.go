package tool //nolint:dupl // 与 update_collaborator_role.go 结构相似但逻辑不同

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// AddCollaborator 添加协作者工具.
type AddCollaborator struct {
	repoDetailClient   RepoDetailClient
	collaboratorClient CollaboratorClient
}

// NewAddCollaborator 创建 AddCollaborator 工具.
func NewAddCollaborator(rdc RepoDetailClient, cc CollaboratorClient) *AddCollaborator {
	return &AddCollaborator{
		repoDetailClient:   rdc,
		collaboratorClient: cc,
	}
}

// Info 返回工具元信息.
func (a *AddCollaborator) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "add_collaborator",
		Desc: "向知识库添加协作者。先通过 search_users 找到用户 ID，再用此工具添加。务必先用 ask_user 让用户确认后再执行。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamRepoID: {
				Type:     TypeString,
				Desc:     DescRepoID,
				Required: true,
			},
			ParamUserID: {
				Type:     TypeString,
				Desc:     "用户 ID（通过 search_users 获取）",
				Required: true,
			},
			ParamRole: {
				Type:     TypeString,
				Desc:     "角色：ADMIN、DEVELOPER 或 VIEWER",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (a *AddCollaborator) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
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
	if _, err := a.repoDetailClient.GetRepo(ctx, params.RepoID, userID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := a.collaboratorClient.AddCollaborator(ctx, params.RepoID, params.UserID, params.Role); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "添加协作者失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		KeySuccess:  true,
		ParamRepoID: params.RepoID,
		ParamUserID: params.UserID,
		ParamRole:   params.Role,
	})
	return string(data), nil
}
