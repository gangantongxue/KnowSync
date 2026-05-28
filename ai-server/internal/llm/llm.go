package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	configModel "github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

const systemPrompt = `你是一个知识库助手，帮助用户回答基于其个人知识库中的文章内容的问题。

## 核心原则
1. 始终使用 search_knowledge 工具搜索用户知识库中相关的文章内容
2. 基于搜索到的文章内容回答，并在回答中引用相关文章
3. 如果搜索没有找到相关文章，明确告知用户"未在您的知识库中找到相关文章"，然后根据自身知识回答
4. 对于第一轮对话，使用 update_session_title 工具更新会话标题，使其贴近对话主题
5. 保持回答简洁、准确、有条理`

// ChatModel LLM 聊天模型封装
type ChatModel struct {
	model  *openai.ChatModel
	config *configModel.LLMCfg
}

// NewChatModel 创建 LLM 聊天模型
func NewChatModel(cfg *configModel.LLMCfg) (*ChatModel, error) {
	cm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		ExtraFields: map[string]any{
			"thinking": map[string]any{
				"type":            "enabled",
				"intensity":       cfg.ThinkingIntensity,
				"thinking_tokens": 32000,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	slog.Info("LLM 模型初始化完成", "model", cfg.Model, "base_url", cfg.BaseURL)
	return &ChatModel{model: cm, config: cfg}, nil
}

// GetSystemMessage 返回系统提示消息
func (c *ChatModel) GetSystemMessage() *schema.Message {
	return schema.SystemMessage(systemPrompt)
}

// GetToolModel 获取绑定工具的模型实例（每次调用返回新实例，并发安全）
func (c *ChatModel) GetToolModel(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return c.model.WithTools(tools)
}
