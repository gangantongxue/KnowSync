package handler

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// Handler HTTP 处理器容器，聚合所有依赖以便 Handler 方法使用.
type Handler struct {
	grpcClient  *grpcclient.Client   // gRPC 客户端，用于调用后端服务
	store       *storage.Store       // 本地文件存储，用于保存头像等文件
	authManager *serviceauth.Manager // service token 管理器
}

// NewHandler 创建 Handler 实例.
func NewHandler(grpcClient *grpcclient.Client, store *storage.Store, authManager *serviceauth.Manager) *Handler {
	return &Handler{
		grpcClient:  grpcClient,
		store:       store,
		authManager: authManager,
	}
}

// getAuthUserID 从请求上下文中获取已认证的用户 ID（string）.
func getAuthUserID(ctx *app.RequestContext) (string, bool) {
	uid := ctx.GetString("user_id")
	return uid, uid != ""
}

// parseUserIDParam 从 URL 中获取 user_id 参数.
func parseUserIDParam(ctx *app.RequestContext) (string, error) {
	return ctx.Param("user_id"), nil
}
