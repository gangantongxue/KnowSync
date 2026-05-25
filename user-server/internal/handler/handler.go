package handler

import (
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"

	"github.com/gangantongxue/knowsync/user-server/internal/service"
)

// Handler 处理程序
type Handler struct {
	pb.UnimplementedUserServiceServer
	Service *service.Service
}

// NewHandler 创建处理程序
func NewHandler(service *service.Service) (*Handler, error) {
	return &Handler{Service: service}, nil
}
