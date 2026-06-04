package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// UpdateTitle 更新会话标题工具，实现 Eino InvokableTool 接口.
type UpdateTitle struct {
	titleUpdater SessionTitleUpdater
}

// NewUpdateTitle 创建更新会话标题工具.
func NewUpdateTitle(tu SessionTitleUpdater) *UpdateTitle {
	return &UpdateTitle{titleUpdater: tu}
}

// Info 返回工具元信息.
func (u *UpdateTitle) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "update_session_title",
		Desc: "根据对话内容更新会话标题，使其贴近当前对话的主题。请根据用户的第一条消息总结出最贴合的简短标题。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"title": {
				Type:     TypeString,
				Desc:     "与当前会话内容最贴合的简短标题（不超过50个字）",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (u *UpdateTitle) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return u.execute(ctx, arguments)
}

func (u *UpdateTitle) execute(ctx context.Context, paramsJSON string) (string, error) {
	sessionID, _ := ctx.Value(CtxKeySessionID).(string)

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
