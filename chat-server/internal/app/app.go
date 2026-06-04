// Package app 提供应用初始化和启动逻辑.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/gangantongxue/knowsync/chat-server/internal/handler"
	"github.com/gangantongxue/knowsync/chat-server/internal/repository"
	"github.com/gangantongxue/knowsync/chat-server/internal/service"
	"github.com/gangantongxue/knowsync/chat-server/internal/ws"
	"github.com/gangantongxue/knowsync/chat-server/pkg/config"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"github.com/gangantongxue/knowsync/chat-server/pkg/logger"
	"github.com/gangantongxue/knowsync/chat-server/pkg/redis"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// NewApp 初始化并启动 chat-server 服务
// 流程：加载配置 → 初始化日志 → 数据库 → Redis → 启动 gRPC 和 WebSocket 服务
// 监听 SIGINT/SIGTERM 信号实现优雅退出.
func NewApp() error {
	slog.Info("=====开始初始化应用=====")

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("初始化配置失败", "error", err)
		return err
	}

	logger, err := logger.NewLogger(cfg)
	if err != nil {
		slog.Error("初始化日志记录器失败", "error", err)
		return err
	}

	db, err := database.NewDatabase(cfg, logger)
	if err != nil {
		slog.Error("初始化数据库失败", "error", err)
		return err
	}

	// 自动迁移数据库表结构
	if err := db.DB.AutoMigrate(&schema.FriendRequest{}, &schema.Friend{}, &schema.User{}, &schema.Message{}, &schema.Group{}, &schema.GroupMember{}); err != nil {
		slog.Error("数据库迁移失败", "error", err)
		return err
	}

	// 初始化 WebSocket Hub
	hub := ws.NewHub()

	// 初始化统一的 Repository
	repo := repository.NewRepository(db.DB)

	r, err := redis.NewRedis(cfg, logger)
	if err != nil {
		slog.Error("初始化 Redis 连接失败", "error", err)
		return err
	}

	wsServer, err := ws.NewServer(hub, cfg, r)
	if err != nil {
		slog.Error("初始化 WebSocket 服务失败", "error", err)
		return err
	}

	// 初始化统一的 Service
	svc := service.NewService(repo, hub)

	// 初始化 gRPC Handler
	hdl, err := handler.NewHandler(svc)
	if err != nil {
		slog.Error("初始化 gRPC 处理器失败", "error", err)
		return err
	}

	slog.Info("=====应用初始化完成=====")

	// 启动 gRPC 服务
	addr := fmt.Sprintf(":%d", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("监听端口失败", "address", addr, "error", err)
		return err
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	// 注册 ChatService gRPC 服务
	pb.RegisterChatServiceServer(srv, hdl)

	// 启动 WebSocket 服务（在 goroutine 中运行）
	go func() {
		if err := wsServer.Start(); err != nil {
			slog.Error("WebSocket 服务异常退出", "error", err)
		}
	}()

	// 监听退出信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("正在关闭服务...")

		// 优雅关闭 WebSocket（最多等待 5 秒）
		wsCtx, wsCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer wsCancel()
		if err := wsServer.Shutdown(wsCtx); err != nil {
			slog.Error("关闭 WebSocket 服务失败", "error", err)
		}

		srv.GracefulStop()

		sqlDB, err := db.DB.DB()
		if err == nil {
			sqlDB.Close() //nolint:errcheck,gosec // 关闭数据库连接（fire and forget）
		}

		if err := r.RDB.Close(); err != nil {
			slog.Error("关闭 Redis 连接失败", "error", err)
		}

		slog.Info("服务已关闭")
	}()

	slog.Info("gRPC 服务启动成功", "address", addr)
	return srv.Serve(lis)
}
