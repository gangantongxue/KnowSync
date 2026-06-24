package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
)

// ServiceAuth 内部服务间鉴权中间件，用于 /internal/* 路由
// 支持两种鉴权方式：
//  1. JWT service token（Authorization: Bearer <token>），用于 Chat/Search 等同步请求
//  2. 共享密钥（X-Internal-Secret header），用于异步 Worker 等无 JWT 场景.
func ServiceAuth(authManager *serviceauth.Manager) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 优先尝试共享密钥鉴权
		if internalSecret := string(ctx.GetHeader("X-Internal-Secret")); internalSecret != "" {
			if !authManager.ValidateInternalSecret(internalSecret) {
				ctx.JSON(401, map[string]any{
					"code":    "UNAUTHORIZED",
					"message": "共享密钥无效",
				})
				ctx.Abort()
				return
			}
			ctx.Set("user_id", "")
			ctx.Set("session_id", "")
			ctx.Next(c)
			return
		}

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
