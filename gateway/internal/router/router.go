package router

import (
	"crypto/rsa"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/gangantongxue/knowsync/gateway/internal/handler"
	"github.com/gangantongxue/knowsync/gateway/pkg/config"
	"github.com/gangantongxue/knowsync/gateway/pkg/grpcclient"
	"github.com/gangantongxue/knowsync/gateway/pkg/middleware"
	"github.com/gangantongxue/knowsync/gateway/pkg/serviceauth"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

func Register(h *server.Hertz, cfg *config.Config, grpcClient *grpcclient.Client, store *storage.Store, publicKey *rsa.PublicKey, authManager *serviceauth.Manager) {
	h.Use(middleware.Logging())
	h.Use(middleware.CORS())

	hdl := handler.NewHandler(grpcClient, store, authManager)

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
			users.GET("/:user_id/profile", hdl.GetUserProfile())
			users.GET("/:user_id/repos", hdl.ListUserPublicRepos())
			users.PUT("/:user_id", hdl.UpdateUser())
			users.PUT("/:user_id/avatar", hdl.SetAvatar())
			users.DELETE("/:user_id", hdl.DeleteUser())
			users.PUT("/:user_id/password", hdl.ResetPassword())

			repos := authorized.Group("/repos")
			repos.POST("", hdl.CreateRepo())
			repos.GET("", hdl.ListUserRepos())
			repos.GET("/followed", hdl.ListFollowedRepos())
			repos.GET("/:repo_id", hdl.GetRepo())
			repos.PUT("/:repo_id", hdl.UpdateRepo())
			repos.DELETE("/:repo_id", hdl.DeleteRepo())
			repos.POST("/:repo_id/follow", hdl.FollowRepo())
			repos.DELETE("/:repo_id/follow", hdl.UnfollowRepo())

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

			// 协作者邀请
			invites := authorized.Group("/collaborators")
			invites.POST("/invite", hdl.InviteCollaborator())
			invites.POST("/invitations/:message_id/accept", hdl.AcceptInvitation())
			invites.POST("/invitations/:message_id/reject", hdl.RejectInvitation())

			ai := authorized.Group("/ai")
			ai.POST("/chat", hdl.Chat())
			ai.POST("/search", hdl.Search())
			ai.GET("/sessions", hdl.GetChatSessions())
			ai.GET("/sessions/:session_id/messages", hdl.GetChatMessages())
			ai.DELETE("/sessions/:session_id", hdl.DeleteChatSession())

			// 好友
			friends := authorized.Group("/friends")
			friends.POST("/requests", hdl.SendFriendRequest())
			friends.GET("/requests/received", hdl.GetFriendRequestsByReceiver())
			friends.GET("/requests/sent", hdl.GetFriendRequestsBySender())
			friends.POST("/requests/:request_id/accept", hdl.AcceptFriendRequest())
			friends.POST("/requests/:request_id/reject", hdl.RejectFriendRequest())
			friends.GET("", hdl.GetFriendList())
			friends.DELETE("/:friend_id", hdl.DeleteFriend())
			friends.PUT("/:friend_id/remark", hdl.UpdateFriendRemark())

			// 用户搜索
			authorized.GET("/users/search", hdl.SearchUsers())

			// 消息
			messages := authorized.Group("/messages")
			messages.POST("/private", hdl.SendPrivateMessage())
			messages.POST("/group", hdl.SendGroupMessage())
			messages.GET("", hdl.GetMessages())
			messages.POST("/:message_id/recall", hdl.RecallMessage())
			messages.POST("/forward", hdl.ForwardMessage())
			messages.GET("/unread-count", hdl.GetUnreadCount())

			// 群组
			groups := authorized.Group("/groups")
			groups.POST("", hdl.CreateGroup())
			groups.GET("", hdl.GetUserGroups())
			groups.GET("/:group_id", hdl.GetGroupInfo())
			groups.PUT("/:group_id", hdl.UpdateGroup())
			groups.DELETE("/:group_id", hdl.LeaveGroup())
			groups.POST("/:group_id/members", hdl.AddMembers())
			groups.DELETE("/:group_id/members/:user_id", hdl.RemoveMember())
			groups.GET("/:group_id/members", hdl.GetGroupMembers())
			groups.POST("/:group_id/transfer", hdl.TransferOwnership())
			groups.POST("/:group_id/admins", hdl.SetAdmin())
			groups.DELETE("/:group_id/admins/:user_id", hdl.RemoveAdmin())

			// 会话
			conversations := authorized.Group("/conversations")
			conversations.GET("", hdl.GetConversationList())
			conversations.POST("/:type/:id/read", hdl.MarkConversationRead())
			conversations.POST("/:type/:id/pin", hdl.TogglePin())
			conversations.DELETE("/:type/:id", hdl.DeleteConversation())
		}
	}

	h.GET("/files/*filepath", handler.FileHandler(store))

	internal := h.Group("/internal")
	internal.Use(middleware.ServiceAuth(authManager))
	{
		internal.GET("/file", handler.InternalFileHandler(store))
		internal.GET("/repos/tree", handler.InternalRepoTree(store))
		internal.GET("/repos/user", hdl.InternalListUserRepos())
		internal.GET("/repos/public", hdl.InternalListPublicRepos())
		internal.GET("/repos/detail", hdl.InternalGetRepo())
	}
}
