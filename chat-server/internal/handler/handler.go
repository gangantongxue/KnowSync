// Package handler 提供 gRPC 消息处理逻辑.
package handler

import (
	"github.com/gangantongxue/knowsync/chat-server/internal/service"
	"github.com/gangantongxue/knowsync/chat-server/internal/ws"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// Handler gRPC 处理器，实现所有的 ChatService RPC 接口.
type Handler struct {
	pb.UnimplementedChatServiceServer
	MessageService      *service.MessageService
	ConversationService *service.ConversationService
	FriendService       *service.FriendService
	GroupService        *service.GroupService
	Hub                 *ws.Hub
}

// NewHandler 创建 gRPC 处理器.
func NewHandler(
	msgSvc *service.MessageService,
	convSvc *service.ConversationService,
	friendSvc *service.FriendService,
	groupSvc *service.GroupService,
	hub *ws.Hub,
) *Handler {
	return &Handler{
		MessageService:      msgSvc,
		ConversationService: convSvc,
		FriendService:       friendSvc,
		GroupService:        groupSvc,
		Hub:                 hub,
	}
}
