package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// UpdateRepo 更新知识库信息工具
type UpdateRepo struct {
	repoDetailClient RepoDetailClient
	repoWriteClient  RepoWriteClient
}

// NewUpdateRepo 创建 UpdateRepo 工具
func NewUpdateRepo(rdc RepoDetailClient, rwc RepoWriteClient) *UpdateRepo {
	return &UpdateRepo{
		repoDetailClient: rdc,
		repoWriteClient:  rwc,
	}
}

// Info 返回工具元信息
func (u *UpdateRepo) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_repo",
		Desc: "更新知识库的名称、描述或可见性。如果只是改名称或描述且用户明确说了，可以设置 _skip_confirm: true 跳过确认。注意：修改可见性时系统会强制要求确认。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
			"name": {
				Type:     "string",
				Desc:     "新的知识库名称",
				Required: false,
			},
			"description": {
				Type:     "string",
				Desc:     "新的知识库描述",
				Required: false,
			},
			"visibility": {
				Type:     "string",
				Desc:     "新的可见性：PUBLIC 或 PRIVATE",
				Required: false,
			},
			"_skip_confirm": {
				Type:     "boolean",
				Desc:     "用户已明确确认时设置为 true",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (u *UpdateRepo) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return u.execute(ctx, arguments)
}

func (u *UpdateRepo) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID      string `json:"repo_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{"success": false, "message": "repo_id 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"success": false, "message": "无法获取用户信息"}`, nil
	}

	// 验证仓库存在且有权限
	if _, err := u.repoDetailClient.GetRepo(ctx, params.RepoID, userID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "获取仓库信息失败: %s"}`, err.Error()), nil
	}

	if err := u.repoWriteClient.UpdateRepo(ctx, params.RepoID, userID, params.Name, params.Description, params.Visibility); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "更新知识库失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"success":     true,
		"repo_id":     params.RepoID,
		"name":        params.Name,
		"description": params.Description,
		"visibility":  params.Visibility,
	})
	return string(data), nil
}
