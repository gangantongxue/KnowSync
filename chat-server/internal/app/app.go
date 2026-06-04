package app

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/gangantongxue/knowsync/chat-server/pkg/config"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"github.com/gangantongxue/knowsync/chat-server/pkg/logger"
	"github.com/gangantongxue/knowsync/chat-server/pkg/redis"
)

// NewApp 初始化并启动 chat-server gRPC 服务
// 流程：加载配置 → 初始化日志 → 数据库 → Redis → 启动 gRPC
// 监听 SIGINT/SIGTERM 信号实现优雅退出
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
	if err := db.DB.AutoMigrate(&schema.FriendRequest{}, &schema.Friend{}, &schema.Message{}, &schema.Group{}, &schema.GroupMember{}); err != nil {
		slog.Error("数据库迁移失败", "error", err)
		return err
	}

	r, err := redis.NewRedis(cfg, logger)
	if err != nil {
		slog.Error("初始化 Redis 连接失败", "error", err)
		return err
	}

	slog.Info("=====应用初始化完成=====")

	addr := fmt.Sprintf(":%d", cfg.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("监听端口失败", "address", addr, "error", err)
		return err
	}

	srv := grpc.NewServer()
	reflection.Register(srv)

	// 监听退出信号，实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("正在关闭服务...")

		srv.GracefulStop()

		sqlDB, err := db.DB.DB()
		if err == nil {
			sqlDB.Close()
		}

		if err := r.RDB.Close(); err != nil {
			slog.Error("关闭 Redis 连接失败", "error", err)
		}

		slog.Info("服务已关闭")
	}()

	slog.Info("gRPC 服务启动成功", "address", addr)
	return srv.Serve(lis)
}
