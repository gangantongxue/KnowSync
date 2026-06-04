package handler

import (
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"

	"github.com/gangantongxue/knowsync/repo-server/internal/service"
)

// Handler gRPC 处理程序，实现 pb.RepoServiceServer 接口.
type Handler struct {
	pb.UnimplementedRepoServiceServer
	Service *service.Service
}

// NewHandler 创建 gRPC 处理程序实例.
func NewHandler(svc *service.Service) (*Handler, error) {
	return &Handler{Service: svc}, nil
}
