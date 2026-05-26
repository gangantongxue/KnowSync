package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/gangantongxue/knowsync/gateway/internal/handler"
	"github.com/gangantongxue/knowsync/gateway/pkg/config"
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/middleware"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// Register 注册所有 HTTP 路由及中间件
// 全局中间件：Logging（请求日志）、CORS（跨域）
// /api/v1 分组下，需要认证的路由使用 Auth 中间件
func Register(h *server.Hertz, cfg *config.Config, grpcClient *grpcclient.Client, store *storage.Store) {
	_ = grpcClient
	h.Use(middleware.Logging())
	h.Use(middleware.CORS())

	v1 := h.Group("/api/v1")
	{
		// 无需认证的接口
		auth := v1.Group("/auth")
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)

		v1.POST("/verify-codes", handler.SendVerifyCode)
		v1.POST("/password/forget", handler.ForgetPassword)

		// 需要认证的接口
		authorized := v1.Group("")
		authorized.Use(middleware.Auth(cfg.Auth.JWTSecret))
		{
			authorized.POST("/auth/logout", handler.Logout)

			users := authorized.Group("/users")
			users.GET("/:user_id", handler.GetUser)
			users.PUT("/:user_id", handler.UpdateUser)
			users.PUT("/:user_id/avatar", handler.SetAvatar)
			users.DELETE("/:user_id", handler.DeleteUser)
			users.PUT("/:user_id/password", handler.ResetPassword)
		}
	}

	h.GET("/files/public/*filepath", handler.FileHandler(store))
	h.GET("/files/auth/:token", handler.FileHandler(store))
}
