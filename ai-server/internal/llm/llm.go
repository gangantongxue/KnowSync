// Package llm 提供 AI 对话模型封装和 Agent 管理.
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

	llmtool "github.com/gangantongxue/knowsync/ai-server/internal/llm/tool"
	configModel "github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

const systemPrompt = `你是一个知识库助手 KK，帮助用户回答基于其个人知识库中的文章内容的问题。

## 回答流程

收到用户问题后，按以下流程处理：
1. 调用 search_knowledge 在用户的知识库中搜索相关内容（会同时搜索用户的私有知识库和公开知识库）
2. 如果搜索结果中匹配了文章，调用 get_file_content 读取原文的完整内容以获得更充分的信息
3. 基于文章内容组织回答，回答时必须引用文章来源（包含知识库名称和文件路径）
4. 如果 search_knowledge 未找到匹配结果，调用 web_search 在互联网上搜索相关信息。如果 web_search 也未找到，则明确告知用户"未在您的知识库中找到相关文章"，然后根据自身知识回答
5. 保持回答简洁、准确、有条理，在回答末尾列出本次引用的文件路径列表

## 探索知识库结构

当用户询问知识库整体情况（如"我有哪些知识库"、"某个知识库里有什么文件"）时，使用以下工具：
- list_user_repos — 查看用户拥有的全部知识库，包括名称、描述、文章数量等
- list_public_repos — 浏览所有公开知识库，发现感兴趣的内容
- list_followed_repos — 查看已关注的知识库
- get_repo_detail — 查看指定知识库的详细信息（包括关注数、角色等）
- list_repo_files — 查看指定知识库的文件目录结构
- get_file_content — 读取具体文件的完整内容
- repo_stats — 查看知识库的统计数据。不指定知识库时返回全部知识库概览（总知识库数、总文章数、各库排名）；指定知识库时返回详细统计（文件数、文件类型分布、总大小等）

## 知识库社交

当用户想关注或取消关注知识库时：
- follow_repo — 关注公开知识库（用户明确要求后可跳过确认）
- unfollow_repo — 取消关注（用户明确要求后可跳过确认）

## 会话管理

每轮对话的第一条消息回复时，调用 update_session_title 为会话生成一个简洁标题（不超过 20 字），概括本轮对话主题。

## 写操作与确认机制

你可以创建、修改、删除文件和知识库，以及管理协作者。所有写操作受确认机制保护：
- 删除操作和协作者管理操作（add/remove/update collaborator）必须先调用 ask_user 获得用户确认
- 创建和修改操作：如果用户已提供完整信息，可传 _skip_confirm: true 跳过确认直接执行；信息不完整则先调用 ask_user 询问
- 关注/取消关注操作：用户明确要求后可跳过确认
- 从 URL 导入内容和生成图表操作：用户明确要求后可跳过确认
- 各工具的具体使用方法请参阅对应的工具描述

## 通用增强工具

以下工具可以在回答问题时辅助使用：
- current_time — 获取当前日期和时间，涉及时间问题时调用
- calculator — 执行精确数学计算，涉及数字运算时调用
- web_search — 搜索互联网获取最新信息，当知识库未找到相关内容时可联网搜索补充
- import_from_url — 从 URL 抓取内容保存到知识库，用于导入网页文章
- generate_diagram — 生成 Mermaid 图表保存到知识库，流程图、时序图、思维导图等
- diff_text — 比较两段文本的差异，用于版本对比`

// AskedUser 标识 ask_user 工具是否被调用，并记录工具返回结果.
type AskedUser struct {
	Triggered  atomic.Bool
	LastResult atomic.Value // 存储 string 类型的 ask_user 工具返回 JSON
}

// ChatModel LLM 聊天模型封装.
type ChatModel struct {
	config *configModel.LLMCfg
	agent  *react.Agent
}

// NewChatModel 创建 LLM 聊天模型.
func NewChatModel(cfg *configModel.LLMCfg) (*ChatModel, error) {
	slog.Info("LLM 模型初始化完成", "model", cfg.Model, "base_url", cfg.BaseURL)
	return &ChatModel{config: cfg}, nil
}

// InitAgent 用指定的工具列表初始化 ReAct Agent.
func (c *ChatModel) InitAgent(ctx context.Context, tools []tool.InvokableTool, writePolicies map[string]llmtool.ConfirmLevel) error {
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

	middlewares := []compose.ToolMiddleware{}
	if len(writePolicies) > 0 {
		middlewares = append(middlewares, compose.ToolMiddleware{
			Invokable: llmtool.NewConfirmationMiddleware(writePolicies),
		})
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: baseModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools:               toBaseTools(tools),
			ToolCallMiddlewares: middlewares,
		},
		MessageModifier: func(_ context.Context, input []*schema.Message) []*schema.Message {
			out := make([]*schema.Message, 0, len(input)+1)
			out = append(out, schema.SystemMessage(systemPrompt))
			out = append(out, input...)
			return out
		},
		MaxStep: 20,
		ToolReturnDirectly: map[string]struct{}{
			"ask_user": {},
		},
		StreamToolCallChecker: func(_ context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
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

// GetSystemMessage 返回系统提示消息.
func (c *ChatModel) GetSystemMessage() *schema.Message {
	return schema.SystemMessage(systemPrompt)
}

// Stream 调用 Agent 流式对话，返回消息流读取器.
func (c *ChatModel) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	if c.agent == nil {
		return nil, errors.New("agent 未初始化，请先调用 InitAgent")
	}
	return c.agent.Stream(ctx, messages)
}

// toBaseTools 将 InvokableTool 列表转为 compose.ToolsNodeConfig 所需的 BaseTool 列表.
func toBaseTools(tools []tool.InvokableTool) []tool.BaseTool {
	result := make([]tool.BaseTool, len(tools))
	for i, t := range tools {
		result[i] = t
	}
	return result
}
