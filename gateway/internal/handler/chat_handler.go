package handler

import (
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// ChatHandler 聊天相关 HTTP 处理器.
type ChatHandler struct {
	grpcClient  *grpcclient.Client
	store       *storage.Store
	authManager *serviceauth.Manager
}

// NewChatHandler 创建 ChatHandler 实例.
func NewChatHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *ChatHandler {
	return &ChatHandler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}
