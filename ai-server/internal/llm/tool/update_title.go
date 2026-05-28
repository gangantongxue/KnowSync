package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// UpdateTitle 更新会话标题工具
type UpdateTitle struct {
	Tool
	titleUpdater SessionTitleUpdater
}

// NewUpdateTitle 创建更新会话标题工具
func NewUpdateTitle(tu SessionTitleUpdater) *UpdateTitle {
	return &UpdateTitle{
		titleUpdater: tu,
		Tool: Tool{
			Name:        "update_session_title",
			Description: "根据对话内容更新会话标题，使其贴近当前对话的主题。请根据用户的第一条消息总结出最贴合的简短标题。",
			ParamsJSON:  updateTitleSchema(),
			Handler:     nil,
		},
	}
}

// Init 初始化 Handler（需要在工具被添加到 Set 前调用）
func (u *UpdateTitle) Init(sessionID string) *Tool {
	u.Tool.Handler = func(ctx context.Context, paramsJSON string) (string, error) {
		return u.execute(ctx, sessionID, paramsJSON)
	}
	return &u.Tool
}

func (u *UpdateTitle) execute(ctx context.Context, sessionID, paramsJSON string) (string, error) {
	var params struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析标题参数失败: %w", err)
	}

	title := strings.TrimSpace(params.Title)
	if title == "" {
		return `{"updated": false, "message": "标题不能为空"}`, nil
	}

	// 截断过长标题
	runes := []rune(title)
	if len(runes) > 100 {
		title = string(runes[:100])
	}

	if err := u.titleUpdater.UpdateSessionTitle(sessionID, title); err != nil {
		slog.Error("更新会话标题失败", "session_id", sessionID, "error", err)
		return `{"updated": false, "message": "更新标题失败"}`, nil
	}

	slog.Info("会话标题已更新", "session_id", sessionID, "title", title)
	data, _ := json.Marshal(map[string]any{
		"updated": true,
		"title":   title,
	})
	return string(data), nil
}

func updateTitleSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {
				"type": "string",
				"description": "与当前会话内容最贴合的简短标题（不超过50个字）"
			}
		},
		"required": ["title"]
	}`)
}
