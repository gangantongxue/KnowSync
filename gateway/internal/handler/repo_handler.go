package handler

import (
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// RepoHandler 知识库相关 HTTP 处理器.
type RepoHandler struct {
	grpcClient  *grpcclient.Client
	store       *storage.Store
	authManager *serviceauth.Manager
}

// NewRepoHandler 创建 RepoHandler 实例.
func NewRepoHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *RepoHandler {
	return &RepoHandler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}
