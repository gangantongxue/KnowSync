package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/gangantongxue/knowsync/ai-server/internal/llm/tool"
	"github.com/gangantongxue/knowsync/ai-server/internal/repository"
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

	// 3. 构建对话消息列表
	messages, err := s.buildMessages(ctx, session.ID, message)
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

	// 5. 构建工具集合
	toolSet := s.buildToolSet(ctx, session.ID, reqUserID)

	// 6. 流式对话循环（支持工具调用递归）
	var finalContent, finalThinking strings.Builder

	var askedUser bool

	for range 10 {
		toolModel, err := s.LLM.GetToolModel(toolSet.GetToolInfos())
		if err != nil {
			_ = cb(&ChatEvent{Error: fmt.Errorf("获取工具模型失败: %w", err)})
			return
		}

		done, err := s.streamAndProcess(ctx, toolModel, messages, toolSet, &finalContent, &finalThinking, cb)
		if err != nil {
			_ = cb(&ChatEvent{Error: err})
			return
		}

		if len(done.ToolCalls) > 0 {
			for _, tc := range done.ToolCalls {
				toolResult, isAskUser := s.executeTool(ctx, toolSet, tc, cb)
				askedUser = askedUser || isAskUser
				if toolResult != nil {
					messages = append(messages, toolResult)
				}
			}
			if askedUser {
				break
			}
			continue
		}

		break
	}

	// 7. 保存助手消息（ask_user 时保存带空内容的记录）
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

	// 8. 首次对话自动更新标题（如果模型没有调用 update_session_title）
	titleUpdated := false
	if isFirstRound {
		title := truncateTitle(message)
		if err = s.Repo.UpdateSessionTitle(session.ID, title); err != nil {
			slog.Error("更新会话标题失败", "error", err)
		} else {
			titleUpdated = true
		}
	}

	// 9. 发送完成事件
	_ = cb(&ChatEvent{
		SessionID:    session.ID,
		MessageID:    assistantMsg.ID,
		Finished:     true,
		Title:        session.Title,
		TitleUpdated: titleUpdated,
	})
}

// buildToolSet 构建当前会话的工具集合
func (s *Service) buildToolSet(ctx context.Context, sessionID, userID string) *tool.Set {
	toolSet := tool.NewSet()

	// search_knowledge 工具：语义搜索知识库文章
	sk := tool.NewSearchKnowledge(s.Embedder, s.VectorStore, s.Client, 0)
	toolSet.Register(sk.Init(userID))

	// update_session_title 工具：更新会话标题
	ut := tool.NewUpdateTitle(s.Repo)
	toolSet.Register(ut.Init(sessionID))

	// ask_user 工具：反问用户获取更多信息
	aq := tool.NewAskQuestion()
	toolSet.Register(aq.Init())

	return toolSet
}

// streamAndProcess 流式调用 LLM 并处理输出
func (s *Service) streamAndProcess(
	ctx context.Context,
	toolModel model.ToolCallingChatModel,
	messages []*schema.Message,
	toolSet *tool.Set,
	finalContent, finalThinking *strings.Builder,
	cb ChatCallback,
) (*schema.Message, error) {
	stream, err := toolModel.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}
	defer stream.Close()

	var accumulated *schema.Message

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
			return nil, fmt.Errorf("流式读取失败: %w", err)
		}

		if accumulated == nil {
			accumulated = chunk
		} else {
			accumulated, err = schema.ConcatMessages([]*schema.Message{accumulated, chunk})
			if err != nil {
				slog.Error("合并消息失败", "error", err)
				accumulated = chunk
			}
		}

		// 检测思考 → 内容的转换点
		if state == stateThinking && chunk.ReasoningContent == "" && (chunk.Content != "" || len(chunk.ToolCalls) > 0) {
			state = stateThinkingEnd
			_ = cb(&ChatEvent{ThinkingFinished: true})
		}

		if chunk.ReasoningContent != "" {
			if state == stateNone {
				state = stateThinking
			}
			finalThinking.WriteString(chunk.ReasoningContent)
			_ = cb(&ChatEvent{ThinkingChunk: chunk.ReasoningContent})
		}

		if chunk.Content != "" {
			finalContent.WriteString(chunk.Content)
			_ = cb(&ChatEvent{ContentChunk: chunk.Content})
		}
	}

	if state == stateThinking {
		_ = cb(&ChatEvent{ThinkingFinished: true})
	}

	if accumulated == nil {
		return &schema.Message{}, nil
	}
	return accumulated, nil
}

// executeTool 执行工具调用并返回工具结果消息
// 返回 (message, askedUser) — askedUser 为 true 表示工具触发了反问用户
func (s *Service) executeTool(ctx context.Context, toolSet *tool.Set, tc schema.ToolCall, cb ChatCallback) (*schema.Message, bool) {
	slog.Info("执行工具调用", "tool", tc.Function.Name)

	result, err := toolSet.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
	if err != nil {
		slog.Error("工具执行失败", "tool", tc.Function.Name, "error", err)
		result = fmt.Sprintf(`{"error": "工具执行失败: %s"}`, err.Error())
		return schema.ToolMessage(result, tc.ID, schema.WithToolName(tc.Function.Name)), false
	}

	// 检测 ask_user 动作：需要反问用户
	var action struct {
		Action string `json:"action"`
	}
	if len(result) > 0 && json.Unmarshal([]byte(result), &action) == nil && action.Action == "ask_user" {
		var askData struct {
			Question string   `json:"question"`
			Type     string   `json:"type"`
			Options  []string `json:"options"`
		}
		if json.Unmarshal([]byte(result), &askData) == nil && askData.Question != "" {
			slog.Info("工具触发反问用户", "question", askData.Question, "type", askData.Type, "options", askData.Options)

			opts := make([]AskUserOption, 0, len(askData.Options)+1)
			for _, o := range askData.Options {
				opts = append(opts, AskUserOption{Label: o, Value: o})
			}
			opts = append(opts, AskUserOption{Label: "其他（自定义输入）", Value: "__other__"})

			_ = cb(&ChatEvent{
				AskUser: &AskUserEvent{
					Question: askData.Question,
					Type:     askData.Type,
					Options:  opts,
					HasOther: true,
				},
			})

			return nil, true
		}
	}

	slog.Info("工具执行完成", "tool", tc.Function.Name, "result_length", len(result))
	return schema.ToolMessage(result, tc.ID, schema.WithToolName(tc.Function.Name)), false
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

// buildMessages 构建对话消息列表
func (s *Service) buildMessages(ctx context.Context, sessionID, userMessage string) ([]*schema.Message, error) {
	messages := []*schema.Message{
		s.LLM.GetSystemMessage(),
	}

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

	messages = append(messages, schema.UserMessage(userMessage))
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
