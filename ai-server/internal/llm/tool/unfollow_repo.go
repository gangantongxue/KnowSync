package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type UnfollowRepo struct {
	followClient FollowClient
}

func NewUnfollowRepo(fc FollowClient) *UnfollowRepo {
	return &UnfollowRepo{followClient: fc}
}

func (u *UnfollowRepo) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "unfollow_repo",
		Desc: "取消关注一个知识库。如果用户已明确要求，可以设置 _skip_confirm: true 跳过确认。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "要取消关注的知识库 ID",
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

func (u *UnfollowRepo) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return u.execute(ctx, arguments)
}

func (u *UnfollowRepo) execute(ctx context.Context, paramsJSON string) (string, error) {
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

	if err := u.followClient.UnfollowRepo(ctx, userID, params.RepoID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "取消关注失败: %s"}`, err.Error()), nil
	}

	return `{"success": true, "message": "已取消关注该知识库"}`, nil
}
