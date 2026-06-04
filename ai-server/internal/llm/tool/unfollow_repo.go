package tool //nolint:dupl // 与 follow_repo.go 结构相似但逻辑不同

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// UnfollowRepo 取消关注知识库工具.
type UnfollowRepo struct {
	followClient FollowClient
}

//nolint:revive // self-documenting
func NewUnfollowRepo(fc FollowClient) *UnfollowRepo {
	return &UnfollowRepo{followClient: fc}
}

//nolint:revive // self-documenting
func (u *UnfollowRepo) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "unfollow_repo",
		Desc: "取消关注一个知识库。如果用户已明确要求，可以设置 _skip_confirm: true 跳过确认。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamRepoID: {
				Type:     TypeString,
				Desc:     "要取消关注的知识库 ID",
				Required: true,
			},
			ParamSkipCfm: {
				Type:     TypeBoolean,
				Desc:     DescSkipConfirm,
				Required: false,
			},
		}),
	}, nil
}

//nolint:revive // self-documenting
func (u *UnfollowRepo) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
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
		return ErrRespRepoIDEmpty, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return ErrRespUserInfo, nil
	}

	if err := u.followClient.UnfollowRepo(ctx, userID, params.RepoID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "取消关注失败: %s"}`, err.Error()), nil
	}

	return `{"success": true, "message": "已取消关注该知识库"}`, nil
}
