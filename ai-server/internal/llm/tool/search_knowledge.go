package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/gangantongxue/knowsync/ai-server/internal/vectorstore"
)

// SearchKnowledge 搜索知识库工具，实现 Eino InvokableTool 接口
type SearchKnowledge struct {
	embedder    Embedder
	vectorStore VectorStore
	repoClient  RepoClient
	threshold   float32
}

// NewSearchKnowledge 创建搜索知识库工具
//
// threshold: 余弦相似度阈值，低于该值的搜索结果视为无关（默认 0.5）
func NewSearchKnowledge(embedd Embedder, vs VectorStore, rc RepoClient, threshold float32) *SearchKnowledge {
	if threshold <= 0 {
		threshold = 0.5
	}
	return &SearchKnowledge{
		embedder:    embedd,
		vectorStore: vs,
		repoClient:  rc,
		threshold:   threshold,
	}
}

// Info 返回工具元信息
func (s *SearchKnowledge) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "search_knowledge",
		Desc: "在用户自己的知识库和公开知识库中搜索与问题相关的文章内容。通过语义理解匹配文章，返回最相关的内容片段。用户自己知识库的匹配结果会优先展示。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     "string",
				Desc:     "用户的搜索关键词，从用户问题中提取核心搜索词",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用
func (s *SearchKnowledge) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return s.execute(ctx, arguments)
}

const ownRepoBoost = float32(1.5)

func (s *SearchKnowledge) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析搜索参数失败: %w", err)
	}

	if params.Query == "" {
		return noResultsJSON("搜索关键词为空"), nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	slog.Info("搜索知识库", "query", params.Query, "user_id", userID, "threshold", s.threshold)

	// 1. 获取用户可访问的仓库和公开仓库
	userRepoIDs, err := s.repoClient.ListUserRepos(ctx, userID)
	if err != nil {
		slog.Error("获取用户仓库列表失败", "error", err)
		return noResultsJSON("获取仓库列表失败"), nil
	}

	publicRepoIDs, err := s.repoClient.ListPublicRepos(ctx)
	if err != nil {
		slog.Error("获取公开仓库列表失败", "error", err)
		return noResultsJSON("获取仓库列表失败"), nil
	}

	// 2. 合并仓库列表，标记用户自己的仓库用于权重提升
	ownSet := make(map[string]bool, len(userRepoIDs))
	for _, id := range userRepoIDs {
		ownSet[id] = true
	}

	// 去重：公开仓库中用户已有的不重复加入
	allRepoIDs := make([]string, 0, len(userRepoIDs)+len(publicRepoIDs))
	allRepoIDs = append(allRepoIDs, userRepoIDs...)
	seen := make(map[string]bool, len(allRepoIDs))
	for _, id := range allRepoIDs {
		seen[id] = true
	}
	for _, id := range publicRepoIDs {
		if !seen[id] {
			allRepoIDs = append(allRepoIDs, id)
			seen[id] = true
		}
	}

	if len(allRepoIDs) == 0 {
		slog.Info("没有可搜索的仓库", "user_id", userID)
		return noResultsJSON("未在您的知识库中找到相关文章，将根据自身知识回答"), nil
	}

	// 3. 向量化搜索关键词
	vec64, err := s.embedder.EmbedStrings(ctx, []string{params.Query})
	if err != nil || len(vec64) == 0 {
		slog.Error("向量化查询失败", "error", err)
		return noResultsJSON("搜索处理失败"), nil
	}

	queryEmbedding := make([]float32, len(vec64[0]))
	for i, v := range vec64[0] {
		queryEmbedding[i] = float32(v)
	}

	// 4. 跨仓库搜索（取足够数量以便后续加权排序）
	results, err := s.vectorStore.SearchCrossRepos(ctx, allRepoIDs, queryEmbedding, 20)
	if err != nil {
		slog.Error("搜索向量库失败", "error", err)
		return noResultsJSON("搜索失败"), nil
	}

	// 5. 应用权重：用户自己的仓库分数 × 1.5
	for i, r := range results {
		if ownSet[r.RepoID] {
			results[i].Score = r.Score * ownRepoBoost
		}
	}

	// 6. 按提升后的分数降序重新排序
	for i := range results {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// 7. 按阈值过滤
	var matched []vectorstore.SearchResult
	for _, r := range results {
		if r.Score >= s.threshold {
			matched = append(matched, r)
		}
	}

	if len(matched) == 0 {
		slog.Info("知识库搜索结果均低于阈值", "query", params.Query, "threshold", s.threshold, "total_results", len(results))
		return noResultsJSON("未在您的知识库中找到相关文章，将根据自身知识回答"), nil
	}

	// 5. 格式化结果
	type resultItem struct {
		RepoID  string  `json:"repo_id"`
		Content string  `json:"content"`
		Score   float32 `json:"score"`
	}
	items := make([]resultItem, len(matched))
	for i, r := range matched {
		items[i] = resultItem{
			RepoID:  r.RepoID,
			Content: r.Content,
			Score:   r.Score,
		}
	}

	data, _ := json.Marshal(map[string]any{
		"found":   true,
		"results": items,
	})
	return string(data), nil
}

func noResultsJSON(msg string) string {
	data, _ := json.Marshal(map[string]any{
		"found":   false,
		"message": msg,
		"results": []any{},
	})
	return string(data)
}
