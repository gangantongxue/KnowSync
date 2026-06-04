package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// SearchUsers 搜索用户工具.
type SearchUsers struct {
	userSearchClient UserSearchClient
}

// NewSearchUsers 创建 SearchUsers 工具.
func NewSearchUsers(usc UserSearchClient) *SearchUsers {
	return &SearchUsers{userSearchClient: usc}
}

// Info 返回工具元信息.
func (s *SearchUsers) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_users",
		Desc: "搜索平台上的用户，返回用户列表及 ID。用于查找要添加为协作者的用户。返回结果包含用户 ID 和名称。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"keyword": {
				Type:     TypeString,
				Desc:     "搜索关键词，用于匹配用户名称",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (s *SearchUsers) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return s.execute(ctx, arguments)
}

func (s *SearchUsers) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		Keyword string `json:"keyword"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.Keyword == "" {
		return `{"users": [], "message": "keyword 不能为空"}`, nil
	}

	users, err := s.userSearchClient.SearchUsers(ctx, params.Keyword)
	if err != nil {
		return fmt.Sprintf(`{"users": [], "message": "搜索用户失败: %s"}`, err.Error()), nil
	}

	type userItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	items := make([]userItem, 0, len(users))
	for _, u := range users {
		items = append(items, userItem(u))
	}

	data, _ := json.Marshal(map[string]any{
		"users":  items,
		KeyTotal: len(items),
	})
	return string(data), nil
}
