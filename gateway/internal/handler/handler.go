package handler

import (
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// Handler HTTP 处理器容器，聚合所有依赖以便 Handler 方法使用
type Handler struct {
	grpcClient *grpcclient.Client // gRPC 客户端，用于调用后端服务
	store      *storage.Store     // 本地文件存储，用于保存头像等文件
}

// NewHandler 创建 Handler 实例
func NewHandler(grpcClient *grpcclient.Client, store *storage.Store) *Handler {
	return &Handler{
		grpcClient: grpcClient,
		store:      store,
	}
}
