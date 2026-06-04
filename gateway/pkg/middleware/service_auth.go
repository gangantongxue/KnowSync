package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
)

// ServiceAuth 内部服务间鉴权中间件，用于 /internal/* 路由
// 验证 ai-server 等内部服务发起的请求携带的 service token
func ServiceAuth(authManager *serviceauth.Manager) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authHeader := string(ctx.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(401, map[string]any{
				"code":    "UNAUTHORIZED",
				"message": "缺少 service token",
			})
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		info, err := authManager.Validate(c, tokenString)
		if err != nil {
			ctx.JSON(401, map[string]any{
				"code":    "INVALID_TOKEN",
				"message": err.Error(),
			})
			ctx.Abort()
			return
		}

		ctx.Set("user_id", info.UserID)
		ctx.Set("session_id", info.SessionID)
		ctx.Next(c)
	}
}
