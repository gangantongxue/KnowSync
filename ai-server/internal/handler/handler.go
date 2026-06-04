// Package handler 提供 gRPC 请求处理和响应构造.
package handler

import (
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"

	"github.com/gangantongxue/knowsync/ai-server/internal/service"
)

// Handler gRPC 处理器.
type Handler struct {
	pb.UnimplementedAIServiceServer
	Service *service.Service
}

// NewHandler 创建 gRPC 处理器.
func NewHandler(svc *service.Service) (*Handler, error) {
	return &Handler{Service: svc}, nil
}
