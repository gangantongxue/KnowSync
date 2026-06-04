package tool

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ListFollowedRepos struct {
	followClient FollowClient
}

func NewListFollowedRepos(fc FollowClient) *ListFollowedRepos {
	return &ListFollowedRepos{followClient: fc}
}

func (l *ListFollowedRepos) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "list_followed_repos",
		Desc:        "查看当前用户已关注的知识库列表，包括知识库名称、描述、文章数量等信息。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (l *ListFollowedRepos) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return l.execute(ctx)
}

func (l *ListFollowedRepos) execute(ctx context.Context) (string, error) {
	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{"repos": [], "message": "无法获取用户信息"}`, nil
	}

	repos, err := l.followClient.ListFollowedRepos(ctx, userID)
	if err != nil {
		return `{"repos": [], "message": "获取关注列表失败"}`, nil
	}

	if len(repos) == 0 {
		return `{"repos": [], "message": "你还没有关注任何知识库"}`, nil
	}

	type repoItem struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		OwnerID      string `json:"owner_id"`
		Description  string `json:"description"`
		Visibility   string `json:"visibility"`
		ArticleCount int64  `json:"article_count"`
	}
	items := make([]repoItem, 0, len(repos))
	for _, r := range repos {
		desc := r.Description
		if desc == "" {
			desc = "暂无描述"
		}
		items = append(items, repoItem{
			ID:           r.ID,
			Name:         r.Name,
			OwnerID:      r.OwnerID,
			Description:  desc,
			Visibility:   r.Visibility,
			ArticleCount: r.ArticleCount,
		})
	}

	data, _ := json.Marshal(map[string]any{
		"repos": items,
		"total": len(items),
	})
	return string(data), nil
}
