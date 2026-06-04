// Package handler 提供 gRPC 消息处理逻辑.
package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// GetOnlineStatus 批量查询用户在线状态.
func (h *Handler) GetOnlineStatus(_ context.Context, req *pb.GetOnlineStatusReq) (*pb.GetOnlineStatusResp, error) {
	onlineUsers := h.Service.Hub.GetOnlineUsers(req.GetUserIds())
	onlineStatus := make(map[string]bool, len(req.GetUserIds()))
	for _, userID := range req.GetUserIds() {
		onlineStatus[userID] = false
	}
	for _, userID := range onlineUsers {
		onlineStatus[userID] = true
	}

	return &pb.GetOnlineStatusResp{
		Success:      true,
		Msg:          MsgSuccess,
		OnlineStatus: onlineStatus,
	}, nil
}
