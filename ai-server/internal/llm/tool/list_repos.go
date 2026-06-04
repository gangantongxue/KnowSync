package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ListRepos 查看用户仓库列表工具，实现 Eino InvokableTool 接口.
type ListRepos struct {
	repoDetailClient RepoDetailClient
}

// NewListRepos 创建查看用户仓库列表工具.
func NewListRepos(rdc RepoDetailClient) *ListRepos {
	return &ListRepos{repoDetailClient: rdc}
}

// Info 返回工具元信息.
func (l *ListRepos) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "list_user_repos",
		Desc:        "查看当前用户拥有的全部知识库（仓库）列表，包括仓库名称、描述、可见性、文章数量等信息。当用户询问知识库整体情况或需要了解有哪些知识库时调用此工具。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{}),
	}, nil
}

// InvokableRun 执行工具调用.
func (l *ListRepos) InvokableRun(ctx context.Context, _ string, _ ...tool.Option) (string, error) {
	return l.execute(ctx)
}

func (l *ListRepos) execute(ctx context.Context) (string, error) {
	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return `{KeyRepos: [], "message": "无法获取用户信息"}`, nil
	}

	repos, err := l.repoDetailClient.ListUserReposDetail(ctx, userID)
	if err != nil {
		return fmt.Sprintf(`{KeyRepos: [], "message": "获取仓库列表失败: %s"}`, err.Error()), nil
	}

	if len(repos) == 0 {
		return `{KeyRepos: [], "message": "您还没有创建任何知识库"}`, nil
	}

	type repoItem struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Description  string `json:"description"`
		Visibility   string `json:"visibility"`
		ArticleCount int64  `json:"article_count"`
	}
	items := make([]repoItem, 0, len(repos))
	for _, r := range repos {
		desc := r.Description
		if desc == "" {
			desc = DescNoDesc
		}
		items = append(items, repoItem{
			ID:           r.ID,
			Name:         r.Name,
			Description:  desc,
			Visibility:   r.Visibility,
			ArticleCount: r.ArticleCount,
		})
	}

	data, _ := json.Marshal(map[string]any{
		KeyRepos: items,
		KeyTotal: len(items),
	})
	return string(data), nil
}
