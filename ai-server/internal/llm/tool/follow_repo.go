package tool //nolint:dupl // 与 unfollow_repo.go 结构相似但逻辑不同

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// FollowRepo 关注知识库工具.
type FollowRepo struct {
	followClient FollowClient
}

//nolint:revive // self-documenting
func NewFollowRepo(fc FollowClient) *FollowRepo {
	return &FollowRepo{followClient: fc}
}

//nolint:revive // self-documenting
func (f *FollowRepo) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "follow_repo",
		Desc: "关注一个公开知识库。关注后该知识库的更新会出现在你的关注列表中。如果用户已明确要求，可以设置 _skip_confirm: true 跳过确认。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamRepoID: {
				Type:     TypeString,
				Desc:     "要关注的知识库 ID",
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
func (f *FollowRepo) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
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
		return ErrRespRepoIDEmpty, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return ErrRespUserInfo, nil
	}

	if err := f.followClient.FollowRepo(ctx, userID, params.RepoID); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "关注知识库失败: %s"}`, err.Error()), nil
	}

	return `{"success": true, "message": "已关注该知识库"}`, nil
}
