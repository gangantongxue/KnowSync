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

// ConfirmWriteEvent 写操作确认事件
type ConfirmWriteEvent struct {
	Tool     string `json:"tool"`
	Params   string `json:"params"`
	Question string `json:"question"`
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
	AskUser          *AskUserEvent      // 非空时表示反问用户
	ConfirmWrite     *ConfirmWriteEvent // 非空时表示需要确认写操作
}

// ChatCallback 流式事件回调
type ChatCallback func(event *ChatEvent) error

// Chat 流式对话主逻辑
func (s *Service) Chat(ctx context.Context, reqUserID string, reqSessionID, message string, cb ChatCallback) {
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

	// 9.5 检测 confirm_write 事件（写操作确认）
	confirmWriteEvent := parseConfirmWrite(finalContent.String())
	if confirmWriteEvent != nil {
		question := buildConfirmQuestion(confirmWriteEvent.Tool, confirmWriteEvent.Params)
		confirmWriteEvent.Question = question
		slog.Info("工具触发写操作确认", "tool", confirmWriteEvent.Tool, "question", question)
		_ = cb(&ChatEvent{
			ConfirmWrite: confirmWriteEvent,
		})
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
	} else if confirmWriteEvent != nil {
		assistantMsg.Content = "[系统消息] 等待用户确认操作"
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
func (s *Service) getOrCreateSession(ctx context.Context, userID string, sessionID string) (*repository.ChatSession, error) {
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

// parseConfirmWrite 从助手消息内容中解析 confirm_write 事件
func parseConfirmWrite(content string) *ConfirmWriteEvent {
	idx := strings.Index(content, `"action":"confirm_write"`)
	if idx < 0 {
		idx = strings.Index(content, `"action": "confirm_write"`)
	}
	if idx < 0 {
		return nil
	}

	// 从找到的位置往前找 {
	start := strings.LastIndex(content[:idx], "{")
	if start < 0 {
		return nil
	}

	// 找匹配的 }
	depth := 0
	end := -1
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				goto found
			}
		}
	}
	return nil

found:
	var raw struct {
		Action string `json:"action"`
		Tool   string `json:"tool"`
		Params string `json:"params"`
	}
	if err := json.Unmarshal([]byte(content[start:end]), &raw); err != nil {
		return nil
	}
	if raw.Action != "confirm_write" {
		return nil
	}
	return &ConfirmWriteEvent{
		Tool:   raw.Tool,
		Params: raw.Params,
	}
}

// buildConfirmQuestion 根据工具类型和参数生成确认问题
func buildConfirmQuestion(toolName, paramsJSON string) string {
	var params map[string]any
	json.Unmarshal([]byte(paramsJSON), &params)

	switch toolName {
	case "create_file":
		path, _ := params["file_path"].(string)
		return fmt.Sprintf("确认要在知识库中创建文件 `%s` 吗？", path)
	case "update_file":
		path, _ := params["file_path"].(string)
		return fmt.Sprintf("确认要更新文件 `%s` 的内容吗？", path)
	case "delete_file":
		path, _ := params["file_path"].(string)
		return fmt.Sprintf("⚠️ 确认要永久删除文件 `%s` 吗？此操作不可撤销。", path)
	case "rename_file":
		oldPath, _ := params["old_path"].(string)
		newPath, _ := params["new_path"].(string)
		return fmt.Sprintf("确认要将文件从 `%s` 重命名为 `%s` 吗？", oldPath, newPath)
	case "create_repo":
		name, _ := params["name"].(string)
		return fmt.Sprintf("确认要创建知识库「%s」吗？", name)
	case "update_repo":
		if vis, ok := params["visibility"].(string); ok && vis != "" {
			return fmt.Sprintf("⚠️ 确认要将知识库可见性变更为「%s」吗？", vis)
		}
		return "确认要更新知识库设置吗？"
	case "add_collaborator":
		return "确认要将协作者添加到知识库吗？"
	case "remove_collaborator":
		return "⚠️ 确认要移除该协作者吗？"
	case "update_collaborator_role":
		return "⚠️ 确认要变更协作者角色吗？"
	default:
		return "确认要执行此操作吗？"
	}
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
