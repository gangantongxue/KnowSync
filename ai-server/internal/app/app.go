// Package app 提供应用初始化、依赖注入和生命周期管理.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/gangantongxue/knowsync/ai-server/internal/chunker"
	"github.com/gangantongxue/knowsync/ai-server/internal/embedder"
	"github.com/gangantongxue/knowsync/ai-server/internal/handler"
	"github.com/gangantongxue/knowsync/ai-server/internal/llm"
	"github.com/gangantongxue/knowsync/ai-server/internal/repository"
	"github.com/gangantongxue/knowsync/ai-server/internal/service"
	"github.com/gangantongxue/knowsync/ai-server/internal/vectorstore"
	"github.com/gangantongxue/knowsync/ai-server/internal/worker"
	"github.com/gangantongxue/knowsync/ai-server/pkg/config"
	"github.com/gangantongxue/knowsync/ai-server/pkg/logger"
	"github.com/gangantongxue/knowsync/ai-server/pkg/redis"
)

// NewApp 初始化并启动 ai-server gRPC 服务.
func NewApp() error {
	slog.Info("=====开始初始化应用=====")

	// 1. 加载配置
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("初始化配置失败", "error", err)
		return err
	}

	// 2. 初始化日志记录器
	if _, err := logger.NewLogger(cfg); err != nil {
		slog.Error("初始化日志记录器失败", "error", err)
		return err
	}

	// 3. 连接 Redis
	rdb, err := redis.NewRedis(&cfg.Redis)
	if err != nil {
		slog.Error("连接 Redis 失败", "error", err)
		return err
	}

	// 4. 初始化外部服务客户端
	client, err := service.NewClient(cfg)
	if err != nil {
		slog.Error("初始化外部服务客户端失败", "error", err)
		return err
	}

	// 5. 初始化向量化模型（Eino OpenAI Embedder）
	emb, err := embedder.NewClient(&cfg.Embedder)
	if err != nil {
		slog.Error("初始化向量化模型失败", "error", err)
		return err
	}
	slog.Info("向量化模型客户端初始化完成")

	// 6. 初始化 LLM（DeepSeek ChatModel）
	llmModel, err := llm.NewChatModel(&cfg.LLM)
	if err != nil {
		slog.Error("初始化 LLM 模型失败", "error", err)
		return err
	}

	// 7. 初始化数据库连接
	repo, err := repository.NewRepository(&cfg.Database)
	if err != nil {
		slog.Error("初始化数据库连接失败", "error", err)
		return err
	}

	// 8. 初始化语义切分器
	chunk := chunker.NewChunker(1000)
	slog.Info("语义切分器初始化完成")

	// 9. 初始化向量存储
	vs, err := vectorstore.NewStore(cfg.Chromem.Path)
	if err != nil {
		slog.Error("初始化向量存储失败", "error", err)
		return err
	}

	// 10. 初始化服务层
	svc, err := service.NewService(cfg, rdb, client, repo, llmModel, emb, vs)
	if err != nil {
		slog.Error("初始化服务层失败", "error", err)
		return err
	}

	// 11. 初始化 Worker
	w := worker.NewWorker(rdb, svc, emb, chunk, vs, cfg.Worker.Concurrency, cfg.Worker.MaxRetries)

	// 12. 启动 Worker（后台 goroutine）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Start(ctx)

	// 13. 初始化 gRPC Handler
	hdl, err := handler.NewHandler(svc)
	if err != nil {
		slog.Error("初始化处理器失败", "error", err)
		return err
	}

	// 14. 启动 gRPC 服务器
	addr := fmt.Sprintf(":%d", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("监听端口失败", "address", addr, "error", err)
		return err
	}

	srv := grpc.NewServer()
	pb.RegisterAIServiceServer(srv, hdl)
	reflection.Register(srv)

	slog.Info("=====应用初始化完成=====")

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("正在关闭服务...")
		cancel()

		srv.GracefulStop()

		if err := client.Close(); err != nil {
			slog.Error("关闭外部服务连接失败", "error", err)
		}

		if err := rdb.Close(); err != nil {
			slog.Error("关闭 Redis 连接失败", "error", err)
		}

		slog.Info("服务已关闭")
	}()

	slog.Info("gRPC 服务启动成功", "address", addr)
	return srv.Serve(lis)
}
