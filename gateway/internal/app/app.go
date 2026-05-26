package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/gangantongxue/knowsync/gateway/internal/router"
	"github.com/gangantongxue/knowsync/gateway/pkg/config"
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/logger"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// NewApp 初始化并启动网关服务
// 流程：加载配置 → 初始化日志 → 创建 Hertz 引擎 → 注册路由 → 启动服务
// 监听 SIGINT/SIGTERM 信号实现优雅退出
func NewApp() error {
	slog.Info("=====开始初始化应用=====")

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("初始化配置失败", "error", err)
		return err
	}

	lg, err := logger.NewLogger(cfg)
	if err != nil {
		slog.Error("初始化日志记录器失败", "error", err)
		return err
	}
	_ = lg

	grpcClient, err := grpcclient.NewClient(cfg.GRPCClients.Targets)
	if err != nil {
		slog.Error("初始化 gRPC 客户端失败", "error", err)
		return err
	}

	store, err := storage.NewStore(cfg)
	if err != nil {
		slog.Error("初始化文件存储失败", "error", err)
		return err
	}

	addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
	h := server.New(server.WithHostPorts(addr))

	router.Register(h, cfg, grpcClient, store)

	slog.Info("=====应用初始化完成=====")

	// 监听退出信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		slog.Info("正在关闭服务...")

		if err := grpcClient.Close(); err != nil {
			slog.Error("关闭 gRPC 连接失败", "error", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.Shutdown(ctx)
		slog.Info("服务已关闭")
	}()

	return h.Run()
}
