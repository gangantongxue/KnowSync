package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

const queueKey = "queue:vectorize"

// VectorizeTask 向量化任务
type VectorizeTask struct {
	UserID     string `json:"user_id"`
	RepoID     string `json:"repo_id"`
	FilePath   string `json:"file_path"`
	CreatedAt  string `json:"created_at"`
	RetryCount int    `json:"retry_count"`
}

// VectorizeArticle 将向量化任务推入队列后立即返回
func (s *Service) VectorizeArticle(ctx context.Context, userID, repoID, filePath string) error {
	task := VectorizeTask{
		UserID:    userID,
		RepoID:    repoID,
		FilePath:  filePath,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化任务失败: %w", err)
	}

	if err := s.RDB.LPush(ctx, queueKey, data).Err(); err != nil {
		return fmt.Errorf("任务入队失败: %w", err)
	}

	slog.Info("向量化任务已入队",
		"user_id", userID,
		"repo_id", repoID,
		"file_path", filePath,
	)

	return nil
}
