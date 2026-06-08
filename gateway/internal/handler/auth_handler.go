package handler

import (
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// AuthHandler 认证相关 HTTP 处理器.
type AuthHandler struct {
	grpcClient  *grpcclient.Client
	store       *storage.Store
	authManager *serviceauth.Manager
}

// NewAuthHandler 创建 AuthHandler 实例.
func NewAuthHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *AuthHandler {
	return &AuthHandler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}
