package tool

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ListPublicRepos struct {
	publicRepoClient PublicRepoClient
}

func NewListPublicRepos(prc PublicRepoClient) *ListPublicRepos {
	return &ListPublicRepos{publicRepoClient: prc}
}

func (l *ListPublicRepos) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "list_public_repos",
		Desc:        "浏览所有公开知识库列表，包括知识库名称、描述、文章数量等信息。当用户想发现或探索公开知识库时调用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

func (l *ListPublicRepos) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return l.execute(ctx)
}

func (l *ListPublicRepos) execute(ctx context.Context) (string, error) {
	repos, err := l.publicRepoClient.ListPublicReposDetail(ctx)
	if err != nil {
		return `{"repos": [], "message": "获取公开知识库列表失败"}`, nil
	}

	if len(repos) == 0 {
		return `{"repos": [], "message": "暂无公开知识库"}`, nil
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
