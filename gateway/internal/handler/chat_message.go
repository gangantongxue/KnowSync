package handler

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// SendPrivateMessageReq 发送私聊消息请求体.
type SendPrivateMessageReq struct {
	ReceiverID  string `json:"receiver_id"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
	Extra       string `json:"extra"`
	ReplyToID   string `json:"reply_to_id"`
}

// SendGroupMessageReq 发送群聊消息请求体.
type SendGroupMessageReq struct {
	GroupID     string   `json:"group_id"`
	ContentType string   `json:"content_type"`
	Content     string   `json:"content"`
	Extra       string   `json:"extra"`
	Mentions    []string `json:"mentions"`
	ReplyToID   string   `json:"reply_to_id"`
}

// ForwardMessageReq 转发消息请求体.
type ForwardMessageReq struct {
	TargetConversationType string   `json:"target_conversation_type"`
	TargetConversationID   string   `json:"target_conversation_id"`
	MessageIDs             []string `json:"message_ids"`
}

// SendPrivateMessage 发送私聊消息.
func (h *Handler) SendPrivateMessage() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req SendPrivateMessageReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.SendPrivateMessage(c, &pb.SendPrivateMessageReq{
			SenderId:    uid,
			ReceiverId:  req.ReceiverID,
			ContentType: req.ContentType,
			Content:     req.Content,
			Extra:       req.Extra,
			ReplyToId:   req.ReplyToID,
		})
		if err != nil {
			slog.Error("gRPC SendPrivateMessage 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "发送消息失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			KeyMessage: marshalMessageResp(resp.Message),
		})
	}
}

// SendGroupMessage 发送群聊消息.
func (h *Handler) SendGroupMessage() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req SendGroupMessageReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.SendGroupMessage(c, &pb.SendGroupMessageReq{
			SenderId:    uid,
			GroupId:     req.GroupID,
			ContentType: req.ContentType,
			Content:     req.Content,
			Extra:       req.Extra,
			Mentions:    req.Mentions,
			ReplyToId:   req.ReplyToID,
		})
		if err != nil {
			slog.Error("gRPC SendGroupMessage 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "发送消息失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			KeyMessage: marshalMessageResp(resp.Message),
		})
	}
}

// GetMessages 获取消息列表.
func (h *Handler) GetMessages() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		conversationType := ctx.Query("conversation_type")
		conversationID := ctx.Query("conversation_id")
		if conversationType == "" || conversationID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "会话类型和会话 ID 不能为空")
			return
		}

		var beforeSeqID uint64
		if s := ctx.Query("before_seq_id"); s != "" {
			beforeSeqID, _ = strconv.ParseUint(s, 10, 64)
		}

		limit := int32(20)
		if s := ctx.Query("limit"); s != "" {
			if n, err := strconv.ParseInt(s, 10, 32); err == nil && n > 0 {
				limit = int32(n)
			}
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.GetMessages(c, &pb.GetMessagesReq{
			ConversationType: conversationType,
			ConversationId:   conversationID,
			BeforeSeqId:      beforeSeqID,
			Limit:            limit,
		})
		if err != nil {
			slog.Error("gRPC GetMessages 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取消息列表失败")
			return
		}

		messages := make([]map[string]any, 0, len(resp.Messages))
		for _, m := range resp.Messages {
			messages = append(messages, marshalMessageResp(m))
		}

		response.Success(c, ctx, map[string]any{
			"messages": messages,
		})
	}
}

// RecallMessage 撤回消息.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) RecallMessage() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		messageID := ctx.Param("message_id")
		if messageID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "消息 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		_, err := client.RecallMessage(c, &pb.RecallMessageReq{
			MessageId: messageID,
			SenderId:  uid,
		})
		if err != nil {
			slog.Error("gRPC RecallMessage 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "撤回消息失败")
			return
		}

		response.Success(c, ctx, nil)
	}
}

// ForwardMessage 转发消息.
func (h *Handler) ForwardMessage() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req ForwardMessageReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.ForwardMessage(c, &pb.ForwardMessageReq{
			SenderId:               uid,
			TargetConversationType: req.TargetConversationType,
			TargetConversationId:   req.TargetConversationID,
			MessageIds:             req.MessageIDs,
		})
		if err != nil {
			slog.Error("gRPC ForwardMessage 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "转发消息失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			KeyMessage: marshalMessageResp(resp.Message),
		})
	}
}

// GetUnreadCount 获取未读数.
func (h *Handler) GetUnreadCount() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.GetUnreadCount(c, &pb.GetUnreadCountReq{
			UserId: uid,
		})
		if err != nil {
			slog.Error("gRPC GetUnreadCount 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取未读数失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			"counts": resp.Counts,
		})
	}
}

// marshalMessageResp 将 protobuf Message 转换为 HTTP JSON 响应格式.
func marshalMessageResp(m *pb.Message) map[string]any {
	return map[string]any{
		"id":                m.Id,
		"conversation_type": m.ConversationType,
		"conversation_id":   m.ConversationId,
		"seq_id":            m.SeqId,
		"sender_id":         m.SenderId,
		"content_type":      m.ContentType,
		KeyContent:          m.Content,
		"extra":             m.Extra,
		"reply_to_id":       m.ReplyToId,
		"status":            m.Status,
		KeyCreatedAt:        m.CreatedAt,
	}
}
