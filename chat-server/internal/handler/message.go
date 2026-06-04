package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"

	"github.com/gangantongxue/knowsync/chat-server/internal/service"
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

// SendPrivateMessage 发送私聊消息
func (h *Handler) SendPrivateMessage(ctx context.Context, req *pb.SendPrivateMessageReq) (*pb.SendPrivateMessageResp, error) {
	msg, err := h.Service.SendPrivateMessage(ctx, req.GetSenderId(), req.GetReceiverId(), req.GetContentType(), req.GetContent(), req.GetExtra(), req.GetReplyToId())
	if err != nil {
		return &pb.SendPrivateMessageResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.SendPrivateMessageResp{
		Success: true,
		Msg:     "发送成功",
		Message: messageToPB(msg),
	}, nil
}

// SendGroupMessage 发送群聊消息
func (h *Handler) SendGroupMessage(ctx context.Context, req *pb.SendGroupMessageReq) (*pb.SendGroupMessageResp, error) {
	msg, err := h.Service.SendGroupMessage(ctx, req.GetSenderId(), req.GetGroupId(), req.GetContentType(), req.GetContent(), req.GetExtra(), req.GetMentions(), req.GetReplyToId())
	if err != nil {
		return &pb.SendGroupMessageResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.SendGroupMessageResp{
		Success: true,
		Msg:     "发送成功",
		Message: messageToPB(msg),
	}, nil
}

// GetMessages 获取消息列表
func (h *Handler) GetMessages(ctx context.Context, req *pb.GetMessagesReq) (*pb.GetMessagesResp, error) {
	messages, err := h.Service.GetMessages(ctx, req.GetConversationType(), req.GetConversationId(), req.GetBeforeSeqId(), int(req.GetLimit()))
	if err != nil {
		return &pb.GetMessagesResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	pbMessages := make([]*pb.Message, len(messages))
	for i := range messages {
		pbMessages[i] = messageToPB(&messages[i])
	}
	return &pb.GetMessagesResp{
		Success:  true,
		Msg:      "获取成功",
		Messages: pbMessages,
	}, nil
}

// GetMessageByID 根据ID获取消息
func (h *Handler) GetMessageByID(ctx context.Context, req *pb.GetMessageByIDReq) (*pb.GetMessageByIDResp, error) {
	msg, err := h.Service.Repo.Message.GetMessageByID(ctx, req.GetMessageId())
	if err != nil {
		return &pb.GetMessageByIDResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.GetMessageByIDResp{
		Success: true,
		Msg:     "获取成功",
		Message: messageToPB(msg),
	}, nil
}

// RecallMessage 撤回消息
func (h *Handler) RecallMessage(ctx context.Context, req *pb.RecallMessageReq) (*pb.RecallMessageResp, error) {
	if err := h.Service.RecallMessage(ctx, req.GetMessageId(), req.GetSenderId()); err != nil {
		return &pb.RecallMessageResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.RecallMessageResp{
		Success: true,
		Msg:     "撤回成功",
	}, nil
}

// ForwardMessage 转发消息
func (h *Handler) ForwardMessage(ctx context.Context, req *pb.ForwardMessageReq) (*pb.ForwardMessageResp, error) {
	msg, err := h.Service.ForwardMessage(ctx, req.GetSenderId(), req.GetTargetConversationType(), req.GetTargetConversationId(), req.GetMessageIds())
	if err != nil {
		return &pb.ForwardMessageResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.ForwardMessageResp{
		Success: true,
		Msg:     "转发成功",
		Message: messageToPB(msg),
	}, nil
}

// GetUnreadCount 获取未读数
func (h *Handler) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountReq) (*pb.GetUnreadCountResp, error) {
	convs := req.GetConversation()
	infos := make([]service.ConversationInfo, len(convs))
	for i, c := range convs {
		infos[i] = service.ConversationInfo{
			ConversationType: c.GetConversationType(),
			ConversationID:   c.GetConversationId(),
		}
	}
	counts, err := h.Service.GetUnreadCounts(ctx, req.GetUserId(), infos)
	if err != nil {
		return &pb.GetUnreadCountResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.GetUnreadCountResp{
		Success: true,
		Msg:     "获取成功",
		Counts:  counts,
	}, nil
}
