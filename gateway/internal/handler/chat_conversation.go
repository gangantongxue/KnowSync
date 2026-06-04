package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// GetConversationList 获取会话列表
func (h *Handler) GetConversationList() app.HandlerFunc {
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
		resp, err := client.GetConversationList(c, &pb.GetConversationListReq{
			UserId: uid,
		})
		if err != nil {
			slog.Error("gRPC GetConversationList 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取会话列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		conversations := make([]map[string]any, 0, len(resp.Conversations))
		for _, c := range resp.Conversations {
			conv := map[string]any{
				"conversation_type": c.ConversationType,
				"conversation_id":   c.ConversationId,
				"name":              c.Name,
				"avatar":            c.Avatar,
				"unread_count":      c.UnreadCount,
				"last_message_at":   c.LastMessageAt,
				"pinned":            c.Pinned,
				"mentioned":         c.Mentioned,
			}
			if c.LastMessage != nil {
				conv["last_message"] = marshalMessageResp(c.LastMessage)
			}
			conversations = append(conversations, conv)
		}

		response.Success(c, ctx, map[string]any{
			"conversations": conversations,
		})
	}
}

// MarkConversationRead 标记会话已读
func (h *Handler) MarkConversationRead() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		conversationType := ctx.Param("type")
		conversationID := ctx.Param("id")
		if conversationType == "" || conversationID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "会话类型和会话 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.MarkConversationRead(c, &pb.MarkConversationReadReq{
			UserId:           uid,
			ConversationType: conversationType,
			ConversationId:   conversationID,
		})
		if err != nil {
			slog.Error("gRPC MarkConversationRead 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "标记已读失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// TogglePin 切换会话置顶
func (h *Handler) TogglePin() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		conversationType := ctx.Param("type")
		conversationID := ctx.Param("id")
		if conversationType == "" || conversationID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "会话类型和会话 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.TogglePin(c, &pb.TogglePinReq{
			UserId:           uid,
			ConversationType: conversationType,
			ConversationId:   conversationID,
		})
		if err != nil {
			slog.Error("gRPC TogglePin 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "切换置顶失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"pinned": resp.Pinned,
		})
	}
}

// DeleteConversation 删除会话
func (h *Handler) DeleteConversation() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		conversationType := ctx.Param("type")
		conversationID := ctx.Param("id")
		if conversationType == "" || conversationID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "会话类型和会话 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.DeleteConversation(c, &pb.DeleteConversationReq{
			UserId:           uid,
			ConversationType: conversationType,
			ConversationId:   conversationID,
		})
		if err != nil {
			slog.Error("gRPC DeleteConversation 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "删除会话失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}
