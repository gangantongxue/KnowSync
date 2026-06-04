package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type WebSearch struct {
	searcher WebSearcher
}

func NewWebSearch(ws WebSearcher) *WebSearch {
	return &WebSearch{searcher: ws}
}

func (w *WebSearch) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search",
		Desc: "在互联网上搜索信息。当你的知识储备不足以回答用户问题，或用户询问的是需要最新信息的问题时，使用此工具联网搜索。搜索结果为中文优先。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     "string",
				Desc:     "搜索关键词，从中英文用户问题中提取核心搜索词，使用中文搜索",
				Required: true,
			},
		}),
	}, nil
}

func (w *WebSearch) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return w.execute(ctx, arguments)
}

func (w *WebSearch) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.Query == "" {
		return `{"found": false, "message": "搜索关键词不能为空"}`, nil
	}

	result, err := w.searcher.Search(ctx, params.Query)
	if err != nil {
		return fmt.Sprintf(`{"found": false, "message": "搜索失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		"found":  true,
		"query":  params.Query,
		"result": result,
	})
	return string(data), nil
}
