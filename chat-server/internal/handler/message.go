// Package handler 提供 gRPC 消息处理逻辑.
package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

func messageToPB(m *schema.Message) *pb.Message {
	if m == nil {
		return nil
	}
	pbMsg := &pb.Message{
		Id:               m.ID,
		ConversationType: m.ConversationType,
		ConversationId:   m.ConversationID,
		SeqId:            m.SeqID,
		SenderId:         m.SenderID,
		ContentType:      m.ContentType,
		Content:          m.Content,
		Status:           m.Status,
		CreatedAt:        m.CreatedAt,
	}
	if m.Extra != nil {
		pbMsg.Extra = *m.Extra
	}
	if m.ReplyToID != nil {
		pbMsg.ReplyToId = *m.ReplyToID
	}
	return pbMsg
}

// SendPrivateMessage 发送私聊消息.
func (h *Handler) SendPrivateMessage(ctx context.Context, req *pb.SendPrivateMessageReq) (*pb.SendPrivateMessageResp, error) {
	msg, err := h.MessageService.SendPrivateMessage(ctx, req.GetSenderId(), req.GetReceiverId(), req.GetContentType(), req.GetContent(), req.GetExtra(), req.GetReplyToId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.SendPrivateMessageResp{
		Message: messageToPB(msg),
	}, nil
}

// SendGroupMessage 发送群聊消息.
func (h *Handler) SendGroupMessage(ctx context.Context, req *pb.SendGroupMessageReq) (*pb.SendGroupMessageResp, error) {
	msg, err := h.MessageService.SendGroupMessage(ctx, req.GetSenderId(), req.GetGroupId(), req.GetContentType(), req.GetContent(), req.GetExtra(), req.GetMentions(), req.GetReplyToId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.SendGroupMessageResp{
		Message: messageToPB(msg),
	}, nil
}

// GetMessages 获取消息列表.
func (h *Handler) GetMessages(ctx context.Context, req *pb.GetMessagesReq) (*pb.GetMessagesResp, error) {
	messages, err := h.MessageService.GetMessages(ctx, req.GetConversationType(), req.GetConversationId(), req.GetBeforeSeqId(), int(req.GetLimit()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbMessages := make([]*pb.Message, len(messages))
	for i := range messages {
		pbMessages[i] = messageToPB(&messages[i])
	}
	return &pb.GetMessagesResp{
		Messages: pbMessages,
	}, nil
}

// GetMessageByID 根据ID获取消息.
func (h *Handler) GetMessageByID(ctx context.Context, req *pb.GetMessageByIDReq) (*pb.GetMessageByIDResp, error) {
	msg, err := h.MessageService.GetMessageByID(ctx, req.GetMessageId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetMessageByIDResp{
		Message: messageToPB(msg),
	}, nil
}

// RecallMessage 撤回消息.
func (h *Handler) RecallMessage(ctx context.Context, req *pb.RecallMessageReq) (*pb.RecallMessageResp, error) {
	if err := h.MessageService.RecallMessage(ctx, req.GetMessageId(), req.GetSenderId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RecallMessageResp{}, nil
}

// ForwardMessage 转发消息.
func (h *Handler) ForwardMessage(ctx context.Context, req *pb.ForwardMessageReq) (*pb.ForwardMessageResp, error) {
	msg, err := h.MessageService.ForwardMessage(ctx, req.GetSenderId(), req.GetTargetConversationType(), req.GetTargetConversationId(), req.GetMessageIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ForwardMessageResp{
		Message: messageToPB(msg),
	}, nil
}

// GetUnreadCount 获取未读数.
func (h *Handler) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountReq) (*pb.GetUnreadCountResp, error) {
	counts, err := h.MessageService.GetUserUnreadCounts(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetUnreadCountResp{
		Counts: counts,
	}, nil
}
