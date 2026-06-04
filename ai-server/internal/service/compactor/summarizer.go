package compactor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Summarizer 对话历史摘要生成器
type Summarizer struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// NewSummarizer 创建摘要生成器
func NewSummarizer(baseURL, apiKey, model string) *Summarizer {
	return &Summarizer{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Summarize 生成对话摘要
// existingSummary 为之前已保存的摘要（可为空），messages 为需要摘要的早期消息
func (s *Summarizer) Summarize(ctx context.Context, messages []string, existingSummary string) (string, error) {
	var sb strings.Builder

	sb.WriteString("请将以下对话历史压缩为结构化摘要，保留关键信息。\n\n")
	sb.WriteString("## 用户核心意图\n用户最初想解决什么问题\n\n")
	sb.WriteString("## 已讨论的关键点\n- 重要的事实、决策、确认\n\n")
	sb.WriteString("## 已获取的知识库信息\n- 从知识库检索到的关键内容\n\n")
	sb.WriteString("## 待办/未完成\n- 用户提到但尚未完成的事项\n\n")
	sb.WriteString("请基于以下对话内容填充上述结构：\n\n")

	systemPrompt := sb.String()

	msgs := []chatMessage{
		{Role: "system", Content: systemPrompt},
	}

	if existingSummary != "" {
		msgs = append(msgs, chatMessage{Role: "user", Content: fmt.Sprintf("这是之前已生成的摘要，请在此基础上更新：\n%s\n\n---\n\n以下是需要补充到摘要中的新对话内容：", existingSummary)})
	}

	for _, msg := range messages {
		msgs = append(msgs, chatMessage{Role: "user", Content: msg})
	}

	body := chatRequest{
		Model:    s.model,
		Messages: msgs,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API 调用失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误状态 %d: %s", resp.StatusCode, string(respBody))
	}

	var result chatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API 返回空结果")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
