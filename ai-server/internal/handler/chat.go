package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/metadata"

	"github.com/gangantongxue/knowsync/ai-server/internal/llm/tool"
	"github.com/gangantongxue/knowsync/ai-server/internal/service"
)

// extractServiceToken 从 gRPC metadata 中提取 service token.
func extractServiceToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	tokens := md.Get("x-service-token")
	if len(tokens) == 0 {
		return ""
	}
	return tokens[0]
}

// withServiceToken 将 service token 注入 context.
func withServiceToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, tool.CtxKeyServiceToken, token)
}

// Chat 流式对话.
func (h *Handler) Chat(req *pb.ChatRequest, stream pb.AIService_ChatServer) error {
	cb := func(event *service.ChatEvent) error {
		resp := &pb.ChatResponse{
			SessionId:        event.SessionID,
			MessageId:        event.MessageID,
			ThinkingFinished: event.ThinkingFinished,
			Finished:         event.Finished,
			Title:            event.Title,
			TitleUpdated:     event.TitleUpdated,
		}
		if event.ThinkingChunk != "" {
			resp.Event = &pb.ChatResponse_ThinkingChunk{ThinkingChunk: event.ThinkingChunk}
		}
		if event.ContentChunk != "" {
			resp.Event = &pb.ChatResponse_ContentChunk{ContentChunk: event.ContentChunk}
		}
		if event.AskUser != nil {
			opts := make([]string, len(event.AskUser.Options))
			for i, o := range event.AskUser.Options {
				opts[i] = o.Label
			}
			resp.Event = &pb.ChatResponse_AskUserEvent{
				AskUserEvent: &pb.AskUserEvent{
					Question: event.AskUser.Question,
					Type:     event.AskUser.Type,
					Options:  opts,
					HasOther: event.AskUser.HasOther,
				},
			}
		}
		return stream.Send(resp)
	}

	// 从 gRPC metadata 提取 service token，注入 context 供工具调用使用
	token := extractServiceToken(stream.Context())
	ctx := withServiceToken(stream.Context(), token)
	h.Service.Chat(ctx, req.GetUserId(), req.GetSessionId(), req.GetMessage(), cb)
	return nil
}

// GetChatSessions 获取会话列表（游标分页）.
//
//nolint:revive // ctx required by interface
//nolint:revive // ctx required by interface
func (h *Handler) GetChatSessions(ctx context.Context, req *pb.GetChatSessionsRequest) (*pb.GetChatSessionsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}

	sessions, hasMore, err := h.Service.Repo.ListSessions(req.GetUserId(), req.GetCursor(), limit)
	if err != nil {
		return &pb.GetChatSessionsResponse{ //nolint:nilerr // 项目约定：handler 将业务错误编码到响应体中
			Success: false,
			Msg:     err.Error(),
		}, nil
	}

	pbSessions := make([]*pb.ChatSession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = &pb.ChatSession{
			Id:        s.ID,
			UserId:    s.UserID,
			Title:     s.Title,
			CreatedAt: s.CreatedAt.Unix(),
			UpdatedAt: s.UpdatedAt.Unix(),
		}
	}

	return &pb.GetChatSessionsResponse{
		Success:  true,
		Sessions: pbSessions,
		HasMore:  hasMore,
	}, nil
}

// GetChatMessages 获取会话消息列表（游标分页）.
//
//nolint:revive // ctx required by interface
//nolint:revive // ctx required by interface
func (h *Handler) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}

	messages, hasMore, err := h.Service.Repo.ListMessages(req.GetSessionId(), req.GetCursor(), limit)
	if err != nil {
		return &pb.GetChatMessagesResponse{ //nolint:nilerr // 项目约定：handler 将业务错误编码到响应体中
			Success: false,
			Msg:     err.Error(),
		}, nil
	}

	pbMessages := make([]*pb.ChatMessage, len(messages))
	for i, m := range messages {
		pbMessages[i] = &pb.ChatMessage{
			Id:        m.ID,
			SessionId: m.SessionID,
			Role:      m.Role,
			Content:   m.Content,
			Thinking:  m.Thinking,
			CreatedAt: m.CreatedAt.Unix(),
		}
	}

	return &pb.GetChatMessagesResponse{
		Success:  true,
		Messages: pbMessages,
		HasMore:  hasMore,
	}, nil
}

// DeleteChatSession 删除会话.
//
//nolint:revive // ctx required by interface
//nolint:revive // ctx required by interface
func (h *Handler) DeleteChatSession(ctx context.Context, req *pb.DeleteChatSessionRequest) (*pb.DeleteChatSessionResponse, error) {
	if err := h.Service.Repo.DeleteSession(req.GetSessionId(), req.GetUserId()); err != nil {
		return &pb.DeleteChatSessionResponse{ //nolint:nilerr // 项目约定：handler 将业务错误编码到响应体中
			Success: false,
			Msg:     err.Error(),
		}, nil
	}

	return &pb.DeleteChatSessionResponse{
		Success: true,
	}, nil
}
