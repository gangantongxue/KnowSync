package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type FollowRepo struct {
	followClient FollowClient
}

func NewFollowRepo(fc FollowClient) *FollowRepo {
	return &FollowRepo{followClient: fc}
}

func (f *FollowRepo) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "follow_repo",
		Desc: "关注一个公开知识库。关注后该知识库的更新会出现在你的关注列表中。如果用户已明确要求，可以设置 _skip_confirm: true 跳过确认。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "要关注的知识库 ID",
				Required: true,
			},
			"_skip_confirm": {
				Type:     "boolean",
				Desc:     "用户已明确确认时设置为 true",
				Required: false,
			},
		}),
	}, nil
}

func (f *FollowRepo) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return f.execute(ctx, arguments)
}

func (f *FollowRepo) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID string `json:"repo_id"`
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

	if err := f.followClient.FollowRepo(ctx, userID, params.RepoID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "关注知识库失败: %s"}`, err.Error()), nil
	}

	return `{"success": true, "message": "已关注该知识库"}`, nil
}
