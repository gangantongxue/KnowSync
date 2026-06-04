// Package app 提供应用的初始化、依赖注入和生命周期管理.
package app

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/gangantongxue/knowsync/repo-server/internal/handler"
	"github.com/gangantongxue/knowsync/repo-server/internal/repository"
	"github.com/gangantongxue/knowsync/repo-server/internal/service"
	"github.com/gangantongxue/knowsync/repo-server/pkg/config"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"github.com/gangantongxue/knowsync/repo-server/pkg/logger"
)

// NewApp 创建并启动 gRPC 服务，完成所有依赖的初始化和生命周期管理.
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

	db, err := database.NewDatabase(cfg, lg)
	if err != nil {
		slog.Error("初始化数据库失败", "error", err)
		return err
	}

	// 自动迁移数据库表结构
	if err := db.DB.AutoMigrate(&schema.Repo{}, &schema.Collaborator{}, &schema.Follow{}); err != nil {
		slog.Error("数据库迁移失败", "error", err)
		return err
	}

	repo, err := repository.NewRepository(db)
	if err != nil {
		slog.Error("初始化仓库失败", "error", err)
		return err
	}

	svc, err := service.NewService(repo)
	if err != nil {
		slog.Error("初始化服务失败", "error", err)
		return err
	}

	h, err := handler.NewHandler(svc)
	if err != nil {
		slog.Error("初始化处理程序失败", "error", err)
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
	pb.RegisterRepoServiceServer(srv, h)
	reflection.Register(srv)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		slog.Info("正在关闭服务...")

		srv.GracefulStop()

		sqlDB, err := db.DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				slog.Error("关闭数据库连接失败", "error", err)
			}
		}

		slog.Info("服务已关闭")
	}()

	slog.Info("gRPC 服务启动成功", "address", addr)
	return srv.Serve(lis)
}
