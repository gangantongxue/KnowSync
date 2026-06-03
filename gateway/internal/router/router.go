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

			repos := authorized.Group("/repos")
			repos.POST("", hdl.CreateRepo())
			repos.GET("", hdl.ListUserRepos())
			repos.GET("/:repo_id", hdl.GetRepo())
			repos.PUT("/:repo_id", hdl.UpdateRepo())
			repos.DELETE("/:repo_id", hdl.DeleteRepo())

			files := repos.Group("/:repo_id/files")
			files.GET("/tree", hdl.GetRepoTree())
			files.POST("", hdl.UploadFile())
			files.DELETE("", hdl.DeleteFile())
			files.PUT("", hdl.RenameFile())

			dirs := repos.Group("/:repo_id/dirs")
			dirs.POST("", hdl.MakeDir())

			collabs := repos.Group("/:repo_id/collaborators")
			collabs.POST("", hdl.AddCollaborator())
			collabs.GET("", hdl.ListCollaborators())
			collabs.PUT("/:user_id", hdl.UpdateCollaborator())
			collabs.DELETE("/:user_id", hdl.RemoveCollaborator())

			ai := authorized.Group("/ai")
			ai.POST("/chat", hdl.Chat())
			ai.POST("/search", hdl.Search())
			ai.GET("/sessions", hdl.GetChatSessions())
			ai.GET("/sessions/:session_id/messages", hdl.GetChatMessages())
			ai.DELETE("/sessions/:session_id", hdl.DeleteChatSession())
		}
	}

	h.GET("/files/*filepath", handler.FileHandler(store))

	internal := h.Group("/internal")
	internal.GET("/file", handler.InternalFileHandler(store))
}
