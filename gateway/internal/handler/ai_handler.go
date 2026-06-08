package handler

import (
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// AIHandler AI 相关 HTTP 处理器.
type AIHandler struct {
	grpcClient  *grpcclient.Client
	store       *storage.Store
	authManager *serviceauth.Manager
}

// NewAIHandler 创建 AIHandler 实例.
func NewAIHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *AIHandler {
	return &AIHandler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}
