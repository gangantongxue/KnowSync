// Package handler 提供 gRPC 消息处理逻辑.
package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// GetConversationList 获取会话列表.
func (h *Handler) GetConversationList(ctx context.Context, req *pb.GetConversationListReq) (*pb.GetConversationListResp, error) {
	conversations, err := h.Service.GetConversationList(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	pbConvs := make([]*pb.ConversationInfo, len(conversations))
	for i, c := range conversations {
		pbConvs[i] = &pb.ConversationInfo{
			ConversationType: c.ConversationType,
			ConversationId:   c.ConversationID,
			Name:             c.Name,
			Avatar:           c.Avatar,
			LastMessage:      messageToPB(c.LastMessage),
			UnreadCount:      c.UnreadCount,
			LastMessageAt:    c.LastMessageAt,
			Pinned:           c.Pinned,
			Mentioned:        c.Mentioned,
		}
	}
	return &pb.GetConversationListResp{
		Success:       true,
		Msg:           MsgSuccess,
		Conversations: pbConvs,
	}, nil
}

// MarkConversationRead 标记会话已读.
func (h *Handler) MarkConversationRead(ctx context.Context, req *pb.MarkConversationReadReq) (*pb.MarkConversationReadResp, error) {
	if err := h.Service.MarkConversationRead(ctx, req.GetUserId(), req.GetConversationType(), req.GetConversationId()); err != nil {
		return nil, err
	}
	return &pb.MarkConversationReadResp{
		Success: true,
		Msg:     "已标记已读",
	}, nil
}

// TogglePin 切换置顶.
func (h *Handler) TogglePin(ctx context.Context, req *pb.TogglePinReq) (*pb.TogglePinResp, error) {
	pinned, err := h.Service.TogglePin(ctx, req.GetUserId(), req.GetConversationType(), req.GetConversationId())
	if err != nil {
		return nil, err
	}
	return &pb.TogglePinResp{
		Success: true,
		Msg:     "操作成功",
		Pinned:  pinned,
	}, nil
}

// DeleteConversation 删除会话.
func (h *Handler) DeleteConversation(ctx context.Context, req *pb.DeleteConversationReq) (*pb.DeleteConversationResp, error) {
	if err := h.Service.DeleteConversation(ctx, req.GetUserId(), req.GetConversationType(), req.GetConversationId()); err != nil {
		return nil, err
	}
	return &pb.DeleteConversationResp{
		Success: true,
		Msg:     "删除成功",
	}, nil
}
