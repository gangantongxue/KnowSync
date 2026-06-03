package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/gangantongxue/knowsync/ai-server/internal/repository"
	"github.com/gangantongxue/knowsync/ai-server/internal/llm/tool"
)

// AskUserOption 用户问题选项
type AskUserOption struct {
	Label string
	Value string
}

// AskUserEvent 反问用户事件
type AskUserEvent struct {
	Question string
	Type     string // "single" 单选 / "multiple" 多选
	Options  []AskUserOption
	HasOther bool // 是否包含"其他"自由输入选项
}

// ChatEvent 流式聊天事件
type ChatEvent struct {
	SessionID        string
	MessageID        string
	ThinkingChunk    string
	ContentChunk     string
	ThinkingFinished bool
	Finished         bool
	Title            string
	TitleUpdated     bool
	Error            error
	AskUser          *AskUserEvent // 非空时表示反问用户
}

// ChatCallback 流式事件回调
type ChatCallback func(event *ChatEvent) error

// Chat 流式对话主逻辑
func (s *Service) Chat(ctx context.Context, reqUserID, reqSessionID, message string, cb ChatCallback) {
	// 1. 获取或创建会话
	session, err := s.getOrCreateSession(ctx, reqUserID, reqSessionID)
	if err != nil {
		_ = cb(&ChatEvent{Error: fmt.Errorf("获取会话失败: %w", err)})
		return
	}

	// 2. 保存用户消息
	userMsg := &repository.ChatMessage{
		SessionID: session.ID,
		Role:      "user",
		Content:   message,
	}
	if err = s.Repo.CreateMessage(userMsg); err != nil {
		_ = cb(&ChatEvent{Error: fmt.Errorf("保存用户消息失败: %w", err)})
		return
	}

	// 3. 构建对话消息列表（不含 system prompt，由 agent 的 MessageModifier 注入）
	messages, err := s.buildMessages(ctx, session.ID)
	if err != nil {
		_ = cb(&ChatEvent{Error: fmt.Errorf("构建消息失败: %w", err)})
		return
	}

	// 4. 判断是否为首次对话
	isFirstRound := false
	{
		var msgCount int64
		s.Repo.DB.Model(&repository.ChatMessage{}).Where("session_id = ?", session.ID).Count(&msgCount)
		isFirstRound = msgCount <= 1
	}

	// 5. 重置 ask_user 标记（Agent 复用，每次对话开始时清零）
	if s.AskedUser != nil {
		s.AskedUser.Triggered.Store(false)
		s.AskedUser.LastResult.Store("")
	}

	// 6. 注入请求级上下文，供工具读取
	agentCtx := context.WithValue(ctx, tool.CtxKeyUserID, reqUserID)
	agentCtx = context.WithValue(agentCtx, tool.CtxKeySessionID, session.ID)

	// 7. 调用 Agent 流式对话（Agent 内部自动处理全部工具调用闭环）
	stream, err := s.LLM.Stream(agentCtx, messages)
	if err != nil {
		_ = cb(&ChatEvent{Error: fmt.Errorf("Agent 调用失败: %w", err)})
		return
	}
	defer stream.Close()

	// 8. 处理流式输出
	var finalContent, finalThinking strings.Builder

	const (
		stateNone        = 0
		stateThinking    = 1
		stateThinkingEnd = 2
	)
	state := stateNone

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_ = cb(&ChatEvent{Error: fmt.Errorf("流式读取失败: %w", err)})
			return
		}

		// 检测思考 → 内容的转换点
		if state == stateThinking && chunk.ReasoningContent == "" && (chunk.Content != "" || len(chunk.ToolCalls) > 0) {
			state = stateThinkingEnd
			_ = cb(&ChatEvent{ThinkingFinished: true})
		}

		// 思考内容
		if chunk.ReasoningContent != "" {
			if state == stateNone {
				state = stateThinking
			}
			finalThinking.WriteString(chunk.ReasoningContent)
			_ = cb(&ChatEvent{ThinkingChunk: chunk.ReasoningContent})
		}

		// 正文内容
		if chunk.Content != "" {
			finalContent.WriteString(chunk.Content)
			_ = cb(&ChatEvent{ContentChunk: chunk.Content})
		}
	}

	if state == stateThinking {
		_ = cb(&ChatEvent{ThinkingFinished: true})
	}

	// 9. 处理反问用户场景
	askedUser := false
	if s.AskedUser != nil && s.AskedUser.Triggered.Load() {
		askedUser = true
		lastResult, _ := s.AskedUser.LastResult.Load().(string)
		// 解析 ask_user 工具返回的 JSON，构建反问事件
		var askResult struct {
			Action   string   `json:"action"`
			Question string   `json:"question"`
			Type     string   `json:"type"`
			Options  []string `json:"options"`
		}
		if err := json.Unmarshal([]byte(lastResult), &askResult); err == nil && askResult.Question != "" {
			slog.Info("工具触发反问用户", "question", askResult.Question, "type", askResult.Type, "options", askResult.Options)
			opts := make([]AskUserOption, len(askResult.Options))
			for i, o := range askResult.Options {
				opts[i] = AskUserOption{Label: o, Value: o}
			}
			// HasOther: true 由前端负责添加一个自定义输入选项，避免与工具返回的选项重复
			_ = cb(&ChatEvent{
				AskUser: &AskUserEvent{
					Question: askResult.Question,
					Type:     askResult.Type,
					Options:  opts,
					HasOther: true,
				},
			})
		}
	}

	// 10. 保存助手消息
	assistantMsg := &repository.ChatMessage{
		SessionID: session.ID,
		Role:      "assistant",
		Content:   finalContent.String(),
		Thinking:  finalThinking.String(),
	}
	if askedUser {
		assistantMsg.Content = "[系统消息] 已向用户提问，等待回答"
	}
	if err = s.Repo.CreateMessage(assistantMsg); err != nil {
		slog.Error("保存助手消息失败", "error", err)
	}

	// 11. 首次对话自动更新标题（如果模型没有调用 update_session_title）
	titleUpdated := false
	if isFirstRound {
		title := truncateTitle(message)
		if err = s.Repo.UpdateSessionTitle(session.ID, title); err != nil {
			slog.Error("更新会话标题失败", "error", err)
		} else {
			titleUpdated = true
		}
	}

	// 12. 发送完成事件
	_ = cb(&ChatEvent{
		SessionID:    session.ID,
		MessageID:    assistantMsg.ID,
		Finished:     true,
		Title:        session.Title,
		TitleUpdated: titleUpdated,
	})
}

// getOrCreateSession 获取或创建会话
func (s *Service) getOrCreateSession(ctx context.Context, userID, sessionID string) (*repository.ChatSession, error) {
	if sessionID != "" {
		session, err := s.Repo.GetSession(sessionID)
		if err == nil && session.UserID == userID {
			return session, nil
		}
		slog.Warn("会话不存在或无权访问，创建新会话", "session_id", sessionID, "user_id", userID)
	}

	session := &repository.ChatSession{
		UserID: userID,
		Title:  "新对话",
	}
	if err := s.Repo.CreateSession(session); err != nil {
		return nil, err
	}
	slog.Info("创建新会话", "session_id", session.ID, "user_id", userID)
	return session, nil
}

// buildMessages 构建对话消息列表（不含 system prompt，由 agent 的 MessageModifier 注入）
func (s *Service) buildMessages(ctx context.Context, sessionID string) ([]*schema.Message, error) {
	var messages []*schema.Message

	history, err := s.Repo.GetSessionMessages(sessionID)
	if err != nil {
		return nil, err
	}

	for _, msg := range history {
		if msg.Role == "user" {
			messages = append(messages, schema.UserMessage(msg.Content))
		} else if msg.Role == "assistant" {
			assistantMsg := schema.AssistantMessage(msg.Content, nil)
			if msg.Thinking != "" {
				assistantMsg.ReasoningContent = msg.Thinking
			}
			messages = append(messages, assistantMsg)
		}
	}
	return messages, nil
}

// truncateTitle 截取消息前 n 个字作为标题
func truncateTitle(msg string) string {
	runes := []rune(strings.TrimSpace(msg))
	maxLen := 30
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	if len(runes) == 0 {
		return "新对话"
	}
	return string(runes)
}
