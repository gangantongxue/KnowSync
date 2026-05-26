package app

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/gangantongxue/knowsync/user-server/internal/handler"
	"github.com/gangantongxue/knowsync/user-server/internal/repository"
	"github.com/gangantongxue/knowsync/user-server/internal/service"
	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"github.com/gangantongxue/knowsync/user-server/pkg/database"
	"github.com/gangantongxue/knowsync/user-server/pkg/logger"
	"github.com/gangantongxue/knowsync/user-server/pkg/redis"
)

func NewApp() error {
	slog.Info("=====开始初始化应用=====")
	// 初始化配置
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("初始化配置失败", "error", err)
		return err
	}
	// 初始化日志记录器
	logger, err := logger.NewLogger(cfg)
	if err != nil {
		slog.Error("初始化日志记录器失败", "error", err)
		return err
	}
	// 初始化数据库
	db, err := database.NewDatabase(cfg, logger)
	if err != nil {
		slog.Error("初始化数据库失败", "error", err)
		return err
	}
	// 初始化 Redis 连接
	r, err := redis.NewRedis(cfg, logger)
	if err != nil {
		slog.Error("初始化 Redis 连接失败", "error", err)
		return err
	}
	// 初始化仓库
	repo, err := repository.NewRepository(db, r)
	if err != nil {
		slog.Error("初始化仓库失败", "error", err)
		return err
	}
	// 初始化服务
	service, err := service.NewService(repo)
	if err != nil {
		slog.Error("初始化服务失败", "error", err)
		return err
	}
	// 初始化处理程序
	handler, err := handler.NewHandler(service)
	if err != nil {
		slog.Error("初始化处理程序失败", "error", err)
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
	pb.RegisterUserServiceServer(srv, handler)
	reflection.Register(srv)

	slog.Info("gRPC 服务启动成功", "address", addr)
	return srv.Serve(lis)
}
