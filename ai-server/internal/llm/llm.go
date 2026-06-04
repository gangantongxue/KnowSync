package llm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	configModel "github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

const systemPrompt = `你是一个知识库助手，帮助用户回答基于其个人知识库中的文章内容的问题。

## 核心原则
1. 始终使用 search_knowledge 工具搜索用户知识库中相关的文章内容，会同时搜索用户的私有知识库和公开知识库，用户自己知识库的匹配结果会优先展示
2. 基于搜索到的文章内容回答，并在回答中引用相关文章
3. 如果 search_knowledge 没有搜索到相关文章，明确告知用户"未在您的知识库中找到相关文章"，然后根据自身知识回答
4. 对于第一轮对话，使用 update_session_title 工具更新会话标题，使其贴近对话主题
5. 使用 list_user_repos 工具查看用户有哪些知识库，使用 list_repo_files 查看知识库的文件结构，使用 get_file_content 读取文件内容
6. 保持回答简洁、准确、有条理`

// AskedUser 标识 ask_user 工具是否被调用，并记录工具返回结果
type AskedUser struct {
	Triggered  atomic.Bool
	LastResult atomic.Value // 存储 string 类型的 ask_user 工具返回 JSON
}

// ChatModel LLM 聊天模型封装
type ChatModel struct {
	config *configModel.LLMCfg
	agent  *react.Agent
}

// NewChatModel 创建 LLM 聊天模型
func NewChatModel(cfg *configModel.LLMCfg) (*ChatModel, error) {
	slog.Info("LLM 模型初始化完成", "model", cfg.Model, "base_url", cfg.BaseURL)
	return &ChatModel{config: cfg}, nil
}

// InitAgent 用指定的工具列表初始化 ReAct Agent
func (c *ChatModel) InitAgent(ctx context.Context, tools []tool.InvokableTool) error {
	baseModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: c.config.BaseURL,
		APIKey:  c.config.APIKey,
		Model:   c.config.Model,
		ExtraFields: map[string]any{
			"thinking": map[string]any{
				"type":            "enabled",
				"intensity":       c.config.ThinkingIntensity,
				"thinking_tokens": 32000,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: baseModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: toBaseTools(tools),
		},
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			out := make([]*schema.Message, 0, len(input)+1)
			out = append(out, schema.SystemMessage(systemPrompt))
			out = append(out, input...)
			return out
		},
		MaxStep: 20,
		ToolReturnDirectly: map[string]struct{}{
			"ask_user": {},
		},
		StreamToolCallChecker: func(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
			defer sr.Close()
			for {
				msg, err := sr.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return false, err
				}
				if len(msg.ToolCalls) > 0 {
					return true, nil
				}
			}
			return false, nil
		},
	})
	if err != nil {
		return fmt.Errorf("创建 ReAct Agent 失败: %w", err)
	}

	c.agent = agent
	return nil
}

// GetSystemMessage 返回系统提示消息
func (c *ChatModel) GetSystemMessage() *schema.Message {
	return schema.SystemMessage(systemPrompt)
}

// Stream 调用 Agent 流式对话，返回消息流读取器
func (c *ChatModel) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	if c.agent == nil {
		return nil, fmt.Errorf("agent 未初始化，请先调用 InitAgent")
	}
	return c.agent.Stream(ctx, messages)
}

// toBaseTools 将 InvokableTool 列表转为 compose.ToolsNodeConfig 所需的 BaseTool 列表
func toBaseTools(tools []tool.InvokableTool) []tool.BaseTool {
	result := make([]tool.BaseTool, len(tools))
	for i, t := range tools {
		result[i] = t
	}
	return result
}
