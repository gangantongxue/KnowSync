package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/ai-server/internal/vectorstore"
)

// SearchKnowledge 搜索知识库工具
type SearchKnowledge struct {
	Tool
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
		Tool: Tool{
			Name:        "search_knowledge",
			Description: "在用户知识库中搜索与问题相关的文章内容。通过语义理解匹配用户的文章，返回最相关的内容片段。仅搜索用户自己的文章或公开文章。",
			ParamsJSON:  searchKnowledgeSchema(),
			Handler:     nil, // 在 Init 中设置
		},
	}
}

// Init 初始化 Handler（需要在工具被添加到 Set 前调用）
// userID 被闭包捕获，用于限制搜索范围为该用户的仓库
func (s *SearchKnowledge) Init(userID string) *Tool {
	s.Tool.Handler = func(ctx context.Context, paramsJSON string) (string, error) {
		return s.execute(ctx, userID, paramsJSON)
	}
	return &s.Tool
}

func (s *SearchKnowledge) execute(ctx context.Context, userID, paramsJSON string) (string, error) {
	var params struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析搜索参数失败: %w", err)
	}

	if params.Query == "" {
		return noResultsJSON("搜索关键词为空"), nil
	}

	slog.Info("搜索知识库", "query", params.Query, "user_id", userID, "threshold", s.threshold)

	// 1. 获取用户可访问的仓库
	repoIDs, err := s.repoClient.ListUserRepos(ctx, userID)
	if err != nil {
		slog.Error("获取用户仓库列表失败", "error", err)
		return noResultsJSON("获取仓库列表失败"), nil
	}

	if len(repoIDs) == 0 {
		slog.Info("用户没有可搜索的仓库", "user_id", userID)
		return noResultsJSON("未在您的知识库中找到相关文章，将根据自身知识回答"), nil
	}

	// 2. 向量化搜索关键词
	vec64, err := s.embedder.EmbedStrings(ctx, []string{params.Query})
	if err != nil || len(vec64) == 0 {
		slog.Error("向量化查询失败", "error", err)
		return noResultsJSON("搜索处理失败"), nil
	}

	queryEmbedding := make([]float32, len(vec64[0]))
	for i, v := range vec64[0] {
		queryEmbedding[i] = float32(v)
	}

	// 3. 跨仓库搜索
	results, err := s.vectorStore.SearchCrossRepos(ctx, repoIDs, queryEmbedding, 20)
	if err != nil {
		slog.Error("搜索向量库失败", "error", err)
		return noResultsJSON("搜索失败"), nil
	}

	// 4. 按阈值过滤
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

func searchKnowledgeSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "用户的搜索关键词，从用户问题中提取核心搜索词"
			}
		},
		"required": ["query"]
	}`)
}
