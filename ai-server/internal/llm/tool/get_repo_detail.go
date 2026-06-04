package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type GetRepoDetail struct {
	detailGetter RepoDetailGetter
}

func NewGetRepoDetail(dg RepoDetailGetter) *GetRepoDetail {
	return &GetRepoDetail{detailGetter: dg}
}

func (g *GetRepoDetail) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_repo_detail",
		Desc: "获取指定知识库的详细信息，包括名称、描述、可见性、文章数量、关注数、当前用户的角色以及是否已关注。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"repo_id": {
				Type:     "string",
				Desc:     "知识库 ID",
				Required: true,
			},
		}),
	}, nil
}

func (g *GetRepoDetail) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return g.execute(ctx, arguments)
}

func (g *GetRepoDetail) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		RepoID string `json:"repo_id"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.RepoID == "" {
		return `{"found": false, "message": "repo_id 不能为空"}`, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"found": false, "message": "无法获取用户信息"}`, nil
	}

	detail, err := g.detailGetter.GetRepoDetail(ctx, params.RepoID, userID)
	if err != nil {
		return fmt.Sprintf(`{"found": false, "message": "获取仓库详情失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"found":          true,
		"id":             detail.ID,
		"name":           detail.Name,
		"description":    detail.Description,
		"visibility":     detail.Visibility,
		"owner_id":       detail.OwnerID,
		"article_count":  detail.ArticleCount,
		"follower_count": detail.FollowerCount,
		"my_role":        detail.MyRole,
		"is_following":   detail.IsFollowing,
	})
	return string(data), nil
}
