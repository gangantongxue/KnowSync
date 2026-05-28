package router

import (
	"crypto/rsa"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/gangantongxue/knowsync/gateway/internal/handler"
	"github.com/gangantongxue/knowsync/gateway/pkg/config"
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/middleware"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

func Register(h *server.Hertz, cfg *config.Config, grpcClient *grpcclient.Client, store *storage.Store, publicKey *rsa.PublicKey) {
	h.Use(middleware.Logging())
	h.Use(middleware.CORS())

	hdl := handler.NewHandler(grpcClient, store)

	v1 := h.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/register", hdl.Register())
		auth.POST("/login", hdl.Login())
		auth.POST("/refresh", hdl.Refresh())

		v1.POST("/verify-codes", hdl.SendVerifyCode())
		v1.POST("/password/forget", hdl.ForgetPassword())

		authorized := v1.Group("")
		authorized.Use(middleware.Auth(publicKey))
		{
			authorized.POST("/auth/logout", hdl.Logout())

			users := authorized.Group("/users")
			users.GET("/:user_id", hdl.GetUser())
			users.PUT("/:user_id", hdl.UpdateUser())
			users.PUT("/:user_id/avatar", hdl.SetAvatar())
			users.DELETE("/:user_id", hdl.DeleteUser())
			users.PUT("/:user_id/password", hdl.ResetPassword())
		}
	}

	h.GET("/files/public/*filepath", handler.FileHandler(store))
	h.GET("/files/auth/:token", handler.FileHandler(store))
}
