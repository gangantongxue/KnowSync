package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"
)

const (
	searchCacheTTL  = 5 * time.Minute
	searchCacheKey  = "search:query:%s"
	defaultPageSize = 20
)

// cachedSearchResult 缓存的搜索结果
type cachedSearchResult struct {
	RepoIDs []string `json:"repo_ids"`
	Total   int      `json:"total"`
}

// pbSearchResponse 避免循环导入 pb 包
type pbSearchResponse struct {
	Success    bool
	Msg        string
	RepoIDs    []string
	TotalPages int32
	HasMore    bool
}

// Search 语义搜索公开知识库
// 按 query 文本哈希缓存，不同用户搜索相同关键词共享同一缓存
func (s *Service) Search(ctx context.Context, query string, page, pageSize int) (*pbSearchResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if query == "" {
		return &pbSearchResponse{
			Success: false,
			Msg:     "搜索关键词不能为空",
		}, nil
	}

	// 1. 尝试从缓存读取
	cacheKey := searchCacheKeyForQuery(query)
	data, err := s.RDB.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var cached cachedSearchResult
		if err := json.Unmarshal(data, &cached); err == nil {
			return paginateResult(&cached, cacheKey, page, pageSize), nil
		}
		slog.Warn("搜索缓存反序列化失败，重新搜索", "error", err)
	}

	// 2. 缓存未命中，执行语义搜索（只搜索公开仓库）
	return s.searchAndCache(ctx, query, page, pageSize, cacheKey)
}

// searchCacheKeyForQuery 根据 query 文本生成缓存键
func searchCacheKeyForQuery(query string) string {
	h := md5.Sum([]byte(query))
	return fmt.Sprintf(searchCacheKey, hex.EncodeToString(h[:]))
}

// paginateResult 从缓存结果中取指定页
func paginateResult(cached *cachedSearchResult, cacheKey string, page, pageSize int) *pbSearchResponse {
	totalPages := int(math.Ceil(float64(cached.Total) / float64(pageSize)))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= cached.Total {
		return &pbSearchResponse{
			Success:    true,
			RepoIDs:    []string{},
			TotalPages: int32(totalPages),
			HasMore:    false,
		}
	}

	if end > cached.Total {
		end = cached.Total
	}

	return &pbSearchResponse{
		Success:    true,
		RepoIDs:    cached.RepoIDs[start:end],
		TotalPages: int32(totalPages),
		HasMore:    end < cached.Total,
	}
}

// searchAndCache 执行搜索并缓存结果
func (s *Service) searchAndCache(ctx context.Context, query string, page, pageSize int, cacheKey string) (*pbSearchResponse, error) {
	// 1. 获取所有公开仓库
	repoIDs, err := s.Client.ListPublicRepos(ctx)
	if err != nil {
		slog.Error("获取公开仓库列表失败", "error", err)
		return &pbSearchResponse{
			Success: false,
			Msg:     "获取仓库列表失败",
		}, nil
	}

	if len(repoIDs) == 0 {
		return &pbSearchResponse{
			Success:    true,
			RepoIDs:    []string{},
			TotalPages: 0,
			HasMore:    false,
		}, nil
	}

	// 2. 向量化搜索关键词
	vec64, err := s.Embedder.EmbedStrings(ctx, []string{query})
	if err != nil || len(vec64) == 0 {
		slog.Error("向量化查询失败", "error", err)
		return &pbSearchResponse{
			Success: false,
			Msg:     "搜索处理失败",
		}, nil
	}

	queryEmbedding := make([]float32, len(vec64[0]))
	for i, v := range vec64[0] {
		queryEmbedding[i] = float32(v)
	}

	// 3. 跨仓库搜索（每个 repo 取 top 5）
	results, err := s.VectorStore.SearchCrossRepos(ctx, repoIDs, queryEmbedding, len(repoIDs)*5)
	if err != nil {
		slog.Error("搜索向量库失败", "error", err)
		return &pbSearchResponse{
			Success: false,
			Msg:     "搜索失败",
		}, nil
	}

	// 4. 去重提取匹配的 repo_id（按相似度降序排列）
	seen := make(map[string]bool)
	var matchedIDs []string
	for _, r := range results {
		if seen[r.RepoID] {
			continue
		}
		seen[r.RepoID] = true
		matchedIDs = append(matchedIDs, r.RepoID)
	}

	totalCount := len(matchedIDs)

	// 5. 缓存完整结果到 Redis（按 query 哈希，不同用户共享）
	cacheData, _ := json.Marshal(cachedSearchResult{
		RepoIDs: matchedIDs,
		Total:   totalCount,
	})
	if err := s.RDB.Set(ctx, cacheKey, cacheData, searchCacheTTL).Err(); err != nil {
		slog.Error("搜索缓存写入失败", "error", err)
	}

	// 6. 返回当前页
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	start := (page - 1) * pageSize
	end := min(start+pageSize, totalCount)

	var pageItems []string
	if start < totalCount {
		pageItems = matchedIDs[start:end]
	} else {
		pageItems = []string{}
	}

	return &pbSearchResponse{
		Success:    true,
		RepoIDs:    pageItems,
		TotalPages: int32(totalPages),
		HasMore:    end < totalCount,
	}, nil
}
