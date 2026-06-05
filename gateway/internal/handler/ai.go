// Package handler provides HTTP request handlers for the gateway.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/metadata"
)

// ChatReq 流式对话请求体.
type ChatReq struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// SearchReq 语义搜索请求体.
type SearchReq struct {
	Query    string `json:"query"`
	Page     int32  `json:"page"`
	PageSize int32  `json:"page_size"`
}

// Chat 流式对话，通过 SSE 返回流式结果.
//
//nolint:gocyclo // 流式对话涉及多事件类型分发，逻辑分支由协议定义决定
func (h *Handler) Chat() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req ChatReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}
		if req.Message == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "消息不能为空")
			return
		}

		conn := h.grpcClient.GetConn("ai_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		// 生成 service token，限制在本次请求链路中生效
		tokenString, jti, err := h.authManager.Generate(c, uid, req.SessionID)
		if err != nil {
			slog.Error("生成 service token 失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务内部错误")
			return
		}

		// 通过 gRPC metadata 传递 service token
		md := metadata.Pairs("x-service-token", tokenString)
		grpcCtx := metadata.NewOutgoingContext(c, md)

		client := pb.NewAIServiceClient(conn)
		stream, err := client.Chat(grpcCtx, &pb.ChatRequest{
			UserId:    uid,
			SessionId: req.SessionID,
			Message:   req.Message,
		})
		if err != nil {
			slog.Error("gRPC Chat 调用失败", "error", err)
			_ = h.authManager.RevokeByJTI(c, jti)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "对话请求失败")
			return
		}

		// 流结束后吊销 token
		defer func() { _ = h.authManager.RevokeByJTI(c, jti) }()

		ctx.Header("Content-Type", "text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")

		for {
			resp, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					slog.Error("SSE 流读取失败", "error", err)
					writeSSEEvent(ctx, "error", map[string]string{
						KeySessionID: "",
						KeyMessage:   "流式响应中断",
					})
				}
				break
			}

			if chunk := resp.GetThinkingChunk(); chunk != "" {
				writeSSEEvent(ctx, "thinking", map[string]string{
					KeySessionID: resp.GetSessionId(),
					KeyContent:   chunk,
				})
			}

			if resp.ThinkingFinished {
				writeSSEEvent(ctx, "thinking_finished", map[string]string{
					KeySessionID: resp.GetSessionId(),
				})
			}

			if chunk := resp.GetContentChunk(); chunk != "" {
				writeSSEEvent(ctx, "content", map[string]string{
					KeySessionID: resp.GetSessionId(),
					KeyContent:   chunk,
				})
			}

			if askUser := resp.GetAskUserEvent(); askUser != nil {
				writeSSEEvent(ctx, "ask_user", map[string]any{
					KeySessionID: resp.GetSessionId(),
					"question":   askUser.Question,
					"type":       askUser.Type,
					"options":    askUser.Options,
					"has_other":  askUser.HasOther,
				})
			}

			if resp.Finished {
				writeSSEEvent(ctx, "done", map[string]any{
					KeySessionID:    resp.GetSessionId(),
					"message_id":    resp.GetMessageId(),
					"title":         resp.GetTitle(),
					"title_updated": resp.TitleUpdated,
				})
				break
			}
		}
	}
}

// writeSSEEvent 写入一条 SSE 事件.
func writeSSEEvent(ctx *app.RequestContext, eventType string, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("SSE 数据序列化失败", "event", eventType, "error", err)
		return
	}
	_, _ = ctx.Write(fmt.Appendf(nil, "event: %s\ndata: %s\n\n", eventType, string(jsonData)))
	_ = ctx.Flush()
}

// Search 语义搜索公开知识库（按 query 缓存，不同用户共享结果）.
func (h *Handler) Search() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req SearchReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}
		if req.Query == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "搜索关键词不能为空")
			return
		}
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}

		conn := h.grpcClient.GetConn("ai_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		tokenString, jti, err := h.authManager.Generate(c, "", "")
		if err != nil {
			slog.Error("生成 service token 失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务内部错误")
			return
		}

		md := metadata.Pairs("x-service-token", tokenString)
		grpcCtx := metadata.NewOutgoingContext(c, md)
		defer func() { _ = h.authManager.RevokeByJTI(c, jti) }()

		client := pb.NewAIServiceClient(conn)
		resp, err := client.Search(grpcCtx, &pb.SearchRequest{
			Query:    req.Query,
			Page:     req.Page,
			PageSize: req.PageSize,
		})
		if err != nil {
			slog.Error("gRPC Search 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "搜索失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 500, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"repo_ids":    resp.RepoIds,
			"total_pages": resp.TotalPages,
			"has_more":    resp.HasMore,
		})
	}
}

// GetChatSessions 获取当前用户的会话列表（游标分页）.
func (h *Handler) GetChatSessions() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		cursor := parseCursorParam(ctx.Query("cursor"))
		limit := parseLimitParam(ctx.Query("limit"), 20)

		conn := h.grpcClient.GetConn("ai_server")

		client := pb.NewAIServiceClient(conn)
		resp, err := client.GetChatSessions(c, &pb.GetChatSessionsRequest{
			UserId: uid,
			Cursor: cursor,
			Limit:  limit,
		})
		if err != nil {
			slog.Error("gRPC GetChatSessions 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取会话列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		sessions := make([]map[string]any, 0, len(resp.Sessions))
		for _, s := range resp.Sessions {
			sessions = append(sessions, marshalSession(s))
		}

		response.Success(c, ctx, map[string]any{
			"sessions": sessions,
			"has_more": resp.HasMore,
		})
	}
}

// GetChatMessages 获取指定会话的消息列表（游标分页）.
func (h *Handler) GetChatMessages() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		sessionID := ctx.Param("session_id")
		if sessionID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "会话 ID 不能为空")
			return
		}

		cursor := parseCursorParam(ctx.Query("cursor"))
		limit := parseLimitParam(ctx.Query("limit"), 50)

		conn := h.grpcClient.GetConn("ai_server")

		client := pb.NewAIServiceClient(conn)
		resp, err := client.GetChatMessages(c, &pb.GetChatMessagesRequest{
			UserId:    uid,
			SessionId: sessionID,
			Cursor:    cursor,
			Limit:     limit,
		})
		if err != nil {
			slog.Error("gRPC GetChatMessages 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取消息列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		messages := make([]map[string]any, 0, len(resp.Messages))
		for _, m := range resp.Messages {
			messages = append(messages, marshalMessage(m))
		}

		response.Success(c, ctx, map[string]any{
			"messages": messages,
			"has_more": resp.HasMore,
		})
	}
}

// DeleteChatSession 删除指定会话.
func (h *Handler) DeleteChatSession() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		sessionID := ctx.Param("session_id")
		if sessionID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "会话 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("ai_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewAIServiceClient(conn)
		resp, err := client.DeleteChatSession(c, &pb.DeleteChatSessionRequest{
			UserId:    uid,
			SessionId: sessionID,
		})
		if err != nil {
			slog.Error("gRPC DeleteChatSession 调用失败", "error", err)
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

// parseCursorParam 解析游标参数，空值或非数字返回 0（从头开始）.
func parseCursorParam(s string) int64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// parseLimitParam 解析每页条数，空值或非数字返回默认值.
func parseLimitParam(s string, defaultVal int32) int32 {
	if s == "" {
		return defaultVal
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return int32(n)
}

// marshalSession 将 protobuf ChatSession 转换为 HTTP JSON 响应格式.
func marshalSession(s *pb.ChatSession) map[string]any {
	return map[string]any{
		"id":         s.Id,
		"title":      s.Title,
		KeyCreatedAt: s.CreatedAt,
		KeyUpdatedAt: s.UpdatedAt,
	}
}

// marshalMessage 将 protobuf ChatMessage 转换为 HTTP JSON 响应格式.
func marshalMessage(m *pb.ChatMessage) map[string]any {
	return map[string]any{
		"id":         m.Id,
		KeyRole:      m.Role,
		KeyContent:   m.Content,
		"thinking":   m.Thinking,
		KeyCreatedAt: m.CreatedAt,
	}
}
