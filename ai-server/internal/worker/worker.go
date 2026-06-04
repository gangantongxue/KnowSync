// Package worker 提供异步任务队列处理能力.
package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gangantongxue/knowsync/ai-server/internal/chunker"
	"github.com/gangantongxue/knowsync/ai-server/internal/embedder"
	"github.com/gangantongxue/knowsync/ai-server/internal/service"
	"github.com/gangantongxue/knowsync/ai-server/internal/vectorstore"
)

const (
	queueKey    = "queue:vectorize"
	poppTimeout = 0 // 0 = 阻塞等待
)

// Worker 后台向量化任务消费者.
type Worker struct {
	rdb         *redis.Client
	svc         *service.Service
	embedder    *embedder.Client
	chunker     *chunker.Chunker
	vectorStore *vectorstore.Store
	sem         chan struct{}
	maxRetries  int
}

// NewWorker 创建 Worker.
func NewWorker(
	rdb *redis.Client,
	svc *service.Service,
	embedder *embedder.Client,
	chunker *chunker.Chunker,
	vectorStore *vectorstore.Store,
	concurrency int,
	maxRetries int,
) *Worker {
	return &Worker{
		rdb:         rdb,
		svc:         svc,
		embedder:    embedder,
		chunker:     chunker,
		vectorStore: vectorStore,
		sem:         make(chan struct{}, concurrency),
		maxRetries:  maxRetries,
	}
}

// Start 启动 worker 循环（阻塞）.
func (w *Worker) Start(ctx context.Context) {
	slog.Info(
		"向量化 Worker 启动",
		"concurrency", cap(w.sem),
		"max_retries", w.maxRetries,
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker 收到退出信号，停止消费")
			return
		default:
			w.processNext(ctx)
		}
	}
}

// processNext 从队列获取并处理下一个任务.
func (w *Worker) processNext(ctx context.Context) {
	result, err := w.rdb.BRPop(ctx, poppTimeout, queueKey).Result()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		slog.Error("从队列获取任务失败", "error", err)
		time.Sleep(time.Second)
		return
	}

	if len(result) < 2 {
		return
	}

	taskJSON := result[1]

	var task service.VectorizeTask
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		slog.Error("解析任务失败", "error", err)
		return
	}

	// 使用信号量控制并发
	w.sem <- struct{}{}
	go func() {
		defer func() { <-w.sem }()
		w.processTask(ctx, &task)
	}()
}

// processTask 处理单个向量化任务.
func (w *Worker) processTask(ctx context.Context, task *service.VectorizeTask) {
	logger := slog.With(
		"user_id", task.UserID,
		"repo_id", task.RepoID,
		"file_path", task.FilePath,
	)
	logger.Info("开始处理向量化任务")

	if task.FilePath == "" {
		logger.Warn("文件路径为空")
		return
	}

	// 1. 从 gateway 获取文章内容
	content, err := w.svc.Client.GetArticleContent(ctx, task.UserID, task.RepoID, task.FilePath)
	if err != nil {
		logger.Error("获取文章内容失败", "error", err)
		w.handleFailure("获取文章内容失败", task, err)
		return
	}

	// 2. 语义切分
	chunks := w.chunker.Split(content)
	if len(chunks) == 0 {
		logger.Warn("文章内容为空")
		return
	}
	logger.Info("文章切分完成", "chunks", len(chunks))

	// 3. 批量向量化
	var chunkTexts []string
	for _, chunk := range chunks {
		chunkTexts = append(chunkTexts, chunk.Text)
	}

	allEmbeddings, err := w.embedder.EmbedStrings(ctx, chunkTexts)
	if err != nil {
		logger.Error("向量化失败", "error", err)
		w.handleFailure("向量化失败", task, err)
		return
	}

	logger.Info("向量化完成", "total_vectors", len(allEmbeddings))

	// 4. 存储到 chromem-go
	if err := w.vectorStore.StoreChunks(ctx, task.RepoID, task.FilePath, chunks, allEmbeddings); err != nil {
		logger.Error("存储向量失败", "error", err)
		w.handleFailure("存储向量失败", task, err)
		return
	}

	logger.Info("向量化任务处理完成")
}

// handleFailure 处理任务失败（重试或丢弃）.
func (w *Worker) handleFailure(reason string, task *service.VectorizeTask, err error) {
	task.RetryCount++
	if task.RetryCount > w.maxRetries {
		slog.Error(
			"任务超过最大重试次数，丢弃",
			"file_path", task.FilePath,
			"retry_count", task.RetryCount,
			"reason", reason,
			"error", err,
		)
		return
	}

	// 重新入队（带指数退避延迟）
	task.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	data, _ := json.Marshal(task)

	if err := w.rdb.RPush(context.Background(), queueKey, data).Err(); err != nil {
		slog.Error("任务重新入队失败", "file_path", task.FilePath, "error", err)
		return
	}

	backoff := time.Duration(1<<task.RetryCount) * time.Second
	slog.Warn(
		"任务重新入队",
		"file_path", task.FilePath,
		"retry_count", task.RetryCount,
		"backoff", backoff,
		"reason", reason,
	)
	time.Sleep(backoff)
}
