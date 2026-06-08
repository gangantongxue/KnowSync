package handler

import (
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// UserHandler 用户相关 HTTP 处理器.
type UserHandler struct {
	grpcClient  *grpcclient.Client
	store       *storage.Store
	authManager *serviceauth.Manager
}

// NewUserHandler 创建 UserHandler 实例.
func NewUserHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *UserHandler {
	return &UserHandler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}
