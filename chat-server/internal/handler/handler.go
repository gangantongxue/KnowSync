// Package handler 提供 gRPC 消息处理逻辑.
package handler

import (
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"

	"github.com/gangantongxue/knowsync/chat-server/internal/service"
)

// Handler gRPC 处理器，实现所有的 ChatService RPC 接口.
type Handler struct {
	pb.UnimplementedChatServiceServer
	Service *service.Service
}

// NewHandler 创建 gRPC 处理器.
func NewHandler(svc *service.Service) (*Handler, error) {
	return &Handler{Service: svc}, nil
}
