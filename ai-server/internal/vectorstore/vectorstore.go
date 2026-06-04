package vectorstore

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/philippgille/chromem-go"

	"github.com/gangantongxue/knowsync/ai-server/internal/chunker"
)

const (
	collectionPrefix = "repo_"
	batchSize        = 50
)

// Store 向量存储封装
type Store struct {
	db   *chromem.DB
	path string
}

// NewStore 创建向量存储
func NewStore(path string) (*Store, error) {
	db, err := chromem.NewPersistentDB(path, false)
	if err != nil {
		return nil, fmt.Errorf("初始化 chromem 失败: %w", err)
	}

	slog.Info("向量存储初始化完成", "path", path)
	return &Store{db: db, path: path}, nil
}

// collectionName 生成 Collection 名称
func collectionName(repoID string) string {
	name := strings.NewReplacer("-", "_", ".", "_", ":", "_").Replace(repoID)
	return collectionPrefix + name
}

// getOrCreateCollection 获取或创建 Collection
func (s *Store) getOrCreateCollection(ctx context.Context, repoID string) (*chromem.Collection, error) {
	name := collectionName(repoID)

	col := s.db.GetCollection(name, nil)
	if col != nil {
		return col, nil
	}

	col, err := s.db.CreateCollection(name, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 collection 失败: %w", err)
	}

	slog.Info("创建向量 Collection", "repo_id", repoID, "collection", name)
	return col, nil
}

// StoreChunks 存储文档块的向量，embeddings 为每个 chunk 对应的向量，filePath 作为唯一标识
func (s *Store) StoreChunks(ctx context.Context, repoID, filePath string, chunks []chunker.Chunk, embeddings [][]float64) error {
	if len(chunks) == 0 {
		return nil
	}
	if len(chunks) != len(embeddings) {
		return fmt.Errorf("chunks 与 embeddings 数量不匹配: %d vs %d", len(chunks), len(embeddings))
	}

	// 先删除该 file_path 的旧向量
	if err := s.DeleteFileVectors(ctx, repoID, filePath); err != nil {
		slog.Warn("删除旧向量失败", "file_path", filePath, "error", err)
	}

	col, err := s.getOrCreateCollection(ctx, repoID)
	if err != nil {
		return err
	}

	totalChunks := len(chunks)

	// 分批添加文档
	for i := 0; i < len(chunks); i += batchSize {
		end := min(i+batchSize, len(chunks))

		var docs []chromem.Document
		for j := i; j < end; j++ {
			// float64 → float32 转换，适配 chromem-go
			vec64 := embeddings[j]
			chunkEmbedding := make([]float32, len(vec64))
			for k, v := range vec64 {
				chunkEmbedding[k] = float32(v)
			}

			docID := fmt.Sprintf("%s_%d", strings.ReplaceAll(filePath, "/", "_"), chunks[j].Index)

			doc := chromem.Document{
				ID:      docID,
				Content: chunks[j].Text,
				Metadata: map[string]string{
					"file_path":    filePath,
					"repo_id":      repoID,
					"chunk_index":  fmt.Sprintf("%d", chunks[j].Index),
					"total_chunks": fmt.Sprintf("%d", totalChunks),
				},
				Embedding: chunkEmbedding,
			}
			docs = append(docs, doc)
		}

		if err := col.AddDocuments(ctx, docs, 1); err != nil {
			return fmt.Errorf("添加文档到向量库失败: %w", err)
		}
	}

	slog.Info("存储向量完成",
		"file_path", filePath,
		"repo_id", repoID,
		"chunks", totalChunks,
	)

	return nil
}

// DeleteFileVectors 删除指定 file_path 的所有向量
func (s *Store) DeleteFileVectors(ctx context.Context, repoID, filePath string) error {
	col := s.db.GetCollection(collectionName(repoID), nil)
	if col == nil {
		return nil
	}

	if err := col.Delete(ctx, map[string]string{"file_path": filePath}, nil); err != nil {
		return fmt.Errorf("删除文件向量失败: %w", err)
	}

	return nil
}

// DeleteRepoVectors 删除指定 repo 的所有向量数据
func (s *Store) DeleteRepoVectors(repoID string) error {
	name := collectionName(repoID)
	if err := s.db.DeleteCollection(name); err != nil {
		slog.Error("删除向量 Collection 失败", "repo_id", repoID, "error", err)
		return fmt.Errorf("删除向量 Collection 失败: %w", err)
	}
	slog.Info("删除向量 Collection 成功", "repo_id", repoID, "collection", name)
	return nil
}

// Search 在指定 repo 中搜索相似内容
func (s *Store) Search(ctx context.Context, repoID string, embedding []float32, limit int) ([]chromem.Result, error) {
	col := s.db.GetCollection(collectionName(repoID), nil)
	if col == nil {
		return nil, nil
	}

	results, err := col.QueryEmbedding(ctx, embedding, limit, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("向量搜索失败: %w", err)
	}

	return results, nil
}

// SearchResult 跨 repo 搜索结果
type SearchResult struct {
	RepoID  string
	Content string
	Score   float32
}

// SearchCrossRepos 在多个 repo 中搜索相似内容，返回合并结果
func (s *Store) SearchCrossRepos(ctx context.Context, repoIDs []string, embedding []float32, limit int) ([]SearchResult, error) {
	var all []SearchResult

	for _, repoID := range repoIDs {
		results, err := s.Search(ctx, repoID, embedding, limit)
		if err != nil {
			slog.Warn("搜索 repo 失败", "repo_id", repoID, "error", err)
			continue
		}
		for _, r := range results {
			all = append(all, SearchResult{
				RepoID:  repoID,
				Content: r.Content,
				Score:   r.Similarity,
			})
		}
	}

	// 按相似度排序（冒泡）
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].Score > all[i].Score {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	// 截取 top N
	if len(all) > limit {
		all = all[:limit]
	}

	return all, nil
}
