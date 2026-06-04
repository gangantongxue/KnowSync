package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// CreateRepo 创建知识库工具
type CreateRepo struct {
	repoWriteClient RepoWriteClient
}

// NewCreateRepo 创建 CreateRepo 工具
func NewCreateRepo(rwc RepoWriteClient) *CreateRepo {
	return &CreateRepo{repoWriteClient: rwc}
}

// Info 返回工具元信息
func (c *CreateRepo) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "create_repo",
		Desc: "创建新的知识库。如果用户已经明确说了名称和描述，可以设置 _skip_confirm: true 跳过确认直接执行。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name": {
				Type:     "string",
				Desc:     "知识库名称",
				Required: true,
			},
			"description": {
				Type:     "string",
				Desc:     "知识库描述",
				Required: false,
			},
			"visibility": {
				Type:     "string",
				Desc:     "可见性：PUBLIC 或 PRIVATE（默认 PRIVATE）",
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
func (c *CreateRepo) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return c.execute(ctx, arguments)
}

func (c *CreateRepo) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.Name == "" {
		return `{"success": false, "message": "name 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"success": false, "message": "无法获取用户信息"}`, nil
	}

	visibility := params.Visibility
	if visibility == "" {
		visibility = "PRIVATE"
	}

	repoID, err := c.repoWriteClient.CreateRepo(ctx, userID, params.Name, params.Description, visibility)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "创建知识库失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"success":     true,
		"repo_id":     repoID,
		"name":        params.Name,
		"description": params.Description,
		"visibility":  visibility,
	})
	return string(data), nil
}
