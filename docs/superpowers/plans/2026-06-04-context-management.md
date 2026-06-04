# LLM 对话上下文管理实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 knowsync 的 AI 知识库问答助手添加智能上下文管理机制（溢出检测 + 摘要压缩），避免对话超长时被 LLM API 截断。

**Architecture:** 在 `service` 层新增 `compactor` 包，封装 token 估算、溢出检测和摘要压缩逻辑。`Chat()` 方法在 `buildMessages()` 之后调用 `Compactor.CompactIfNeeded()`，自动检测并压缩超出上下文阈值的对话历史。

**Tech Stack:** Go, GORM, Eino (schema.Message), DeepSeek Chat API (OpenAI-compatible)

---

### Task 1: ChatSession 模型增加摘要字段

**Files:**
- Modify: `ai-server/internal/repository/session.go:11-17`

- [ ] **Step 1: 修改 ChatSession 结构体，添加 Summary 和 SummaryUpdatedAt 字段**

```go
// ChatSession 会话表
type ChatSession struct {
	ID               string     `gorm:"primaryKey;type:char(20)" json:"id"`
	UserID           string     `gorm:"column:user_id;type:varchar(20);not null;index:idx_user_id" json:"user_id"`
	Title            string     `gorm:"column:title;type:varchar(255);not null;default:'新对话'" json:"title"`
	Summary          string     `gorm:"column:summary;type:longtext" json:"summary"`
	SummaryUpdatedAt *time.Time `gorm:"column:summary_updated_at" json:"summary_updated_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
```

- [ ] **Step 2: 添加 UpdateSummary 方法**

在 `session.go` 末尾追加：

```go
// UpdateSummary 更新会话摘要
func (r *Repository) UpdateSummary(sessionID, summary string) error {
	now := time.Now()
	return r.DB.Model(&ChatSession{}).Where("id = ?", sessionID).
		Updates(map[string]any{
			"summary":            summary,
			"summary_updated_at": &now,
		}).Error
}
```

- [ ] **Step 3: 确认 AutoMigrate 覆盖新字段**

`repository.go:39` 已有 `db.AutoMigrate(&ChatSession{}, &ChatMessage{})`，GORM 会自动添加 `summary` 和 `summary_updated_at` 列，无需修改。

---

### Task 2: Token 估算工具

**Files:**
- Create: `ai-server/internal/service/compactor/tokenizer.go`

- [ ] **Step 1: 创建 tokenizer.go**

```go
package compactor

import (
	"math"
)

const avgCharPerToken = 4

// EstimateTokens 估算文本的 token 数（4 字符 ≈ 1 token）
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return int(math.Ceil(float64(len(text)) / avgCharPerToken))
}

// EstimateMessageTokens 估算单条消息的 token 数（消息体 + 角色/格式开销）
func EstimateMessageTokens(content string) int {
	return EstimateTokens(content) + 10 // 10 token 角色/格式开销
}

// EstimateMessagesTokens 估算多条消息的总 token 数
func EstimateMessagesTokens(contents []string) int {
	total := 0
	for _, c := range contents {
		total += EstimateMessageTokens(c)
	}
	return total
}
```

---

### Task 3: 摘要生成器

**Files:**
- Create: `ai-server/internal/service/compactor/summarizer.go`

- [ ] **Step 1: 创建 summarizer.go**

```go
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
```

---

### Task 4: Compactor 编排器

**Files:**
- Create: `ai-server/internal/service/compactor/compactor.go`

- [ ] **Step 1: 创建 compactor.go**

```go
package compactor

import (
	"context"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/gangantongxue/knowsync/ai-server/internal/repository"
)

const (
	// 溢出检测
	OverflowRatio    = 0.8
	ReservedTokens   = 20_000

	// 摘要压缩
	KeepRecentTurns = 3
	MaxSummaryLen   = 2_000
)

// SummaryRepository 摘要存储接口
type SummaryRepository interface {
	UpdateSummary(sessionID, summary string) error
}

// Compactor 上下文管理器
type Compactor struct {
	llmContextLimit int
	summarizer     *Summarizer
	repo           SummaryRepository
}

// New 创建上下文管理器
func New(llmContextLimit int, summarizer *Summarizer, repo SummaryRepository) *Compactor {
	return &Compactor{
		llmContextLimit: llmContextLimit,
		summarizer:     summarizer,
		repo:           repo,
	}
}

// CompactIfNeeded 检测是否需要压缩，需要则执行
func (c *Compactor) CompactIfNeeded(ctx context.Context, messages []*schema.Message, session *repository.ChatSession) ([]*schema.Message, error) {
	if len(messages) < 3 {
		return messages, nil
	}

	threshold := int(float64(c.llmContextLimit-ReservedTokens) * OverflowRatio)
	totalTokens := estimateMessagesTokenCount(messages)

	if totalTokens < threshold {
		return messages, nil
	}

	slog.Info("上下文即将溢出，开始压缩",
		"session_id", session.ID,
		"total_tokens", totalTokens,
		"threshold", threshold,
	)

	return c.compact(ctx, messages, session)
}

// compact 执行压缩：保留最近 N 轮，将早期消息摘要
func (c *Compactor) compact(ctx context.Context, messages []*schema.Message, session *repository.ChatSession) ([]*schema.Message, error) {
	// 将消息按轮分组（user + assistant 为一轮）
	turns := groupTurns(messages)

	if len(turns) <= KeepRecentTurns {
		return messages, nil
	}

	// 分离早期消息和近期消息
	earlyTurns := turns[:len(turns)-KeepRecentTurns]
	recentTurns := turns[len(turns)-KeepRecentTurns:]

	// 收集早期消息的文本用于摘要
	var earlyContents []string
	for _, t := range earlyTurns {
		for _, msg := range t {
			earlyContents = append(earlyContents, msg.Content)
		}
	}

	// 生成摘要
	summary, err := c.summarizer.Summarize(ctx, earlyContents, session.Summary)
	if err != nil {
		slog.Error("生成摘要失败，回退到仅保留最近消息", "error", err)
		// 回退：丢弃早期消息，只保留近期
		return flattenTurns(recentTurns), nil
	}

	// 限制摘要长度
	if len(summary) > MaxSummaryLen {
		summary = string([]rune(summary)[:MaxSummaryLen]) + "..."
	}

	// 持久化摘要
	if err := c.repo.UpdateSummary(session.ID, summary); err != nil {
		slog.Error("保存摘要失败", "error", err)
	}

	slog.Info("上下文压缩完成",
		"session_id", session.ID,
		"early_turns", len(earlyTurns),
		"kept_turns", len(recentTurns),
		"summary_len", len(summary),
	)

	// 组装最终消息：摘要（作为 system 消息）+ 最近消息
	result := []*schema.Message{
		schema.SystemMessage("以下是与该用户的对话历史摘要：\n\n" + summary),
	}
	result = append(result, flattenTurns(recentTurns)...)

	return result, nil
}

// estimateMessagesTokenCount 估算消息列表的总 token 数
func estimateMessagesTokenCount(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		total += EstimateMessageTokens(msg.Content)
	}
	return total
}

// groupTurns 将消息按"用户+助手"轮次分组
// 每条 user 消息开始新的一轮，后续的 assistant 消息归入同一轮
func groupTurns(messages []*schema.Message) [][]*schema.Message {
	var turns [][]*schema.Message
	var current []*schema.Message

	for _, msg := range messages {
		if msg.Role == "user" && len(current) > 0 {
			turns = append(turns, current)
			current = nil
		}
		current = append(current, msg)
	}
	if len(current) > 0 {
		turns = append(turns, current)
	}

	return turns
}

// flattenTurns 将分组消息展平
func flattenTurns(turns [][]*schema.Message) []*schema.Message {
	var result []*schema.Message
	for _, t := range turns {
		result = append(result, t...)
	}
	return result
}
```

---

### Task 5: 集成 Compactor 到 Chat Service

**Files:**
- Modify: `ai-server/internal/service/chat.go:58-83`
- Modify: `ai-server/internal/service/service.go:33-44`

- [ ] **Step 1: Service 结构体增加 Compactor 字段**

修改 `service/service.go`，在 Service 结构体中添加 Compactor 字段：

```go
import (
	"github.com/gangantongxue/knowsync/ai-server/internal/service/compactor"
)

type Service struct {
	Cfg         *model.Config
	RDB         *redis.Client
	Client      *Client
	Web         *WebClient
	Repo        *repository.Repository
	LLM         *llm.ChatModel
	Embedder    *embedder.Client
	VectorStore *vectorstore.Store
	AskedUser   *llm.AskedUser
	Compactor   *compactor.Compactor   // 新增
}
```

- [ ] **Step 2: 在 NewService 中初始化 Compactor**

修改 `service/service.go` 的 `NewService` 函数：

```go
func NewService(cfg *model.Config, rdb *redis.Client, client *Client, repo *repository.Repository, llmModel *llm.ChatModel, emb *embedder.Client, vs *vectorstore.Store) (*Service, error) {
	// 初始化上下文管理器
	summarizer := compactor.NewSummarizer(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)
	ctxMgr := compactor.New(128_000, summarizer, repo)

	svc := &Service{
		Cfg:         cfg,
		RDB:         rdb,
		Client:      client,
		Web:         NewWebClient(&cfg.WebSearch),
		Repo:        repo,
		LLM:         llmModel,
		Embedder:    emb,
		VectorStore: vs,
		AskedUser:   &llm.AskedUser{},
		Compactor:   ctxMgr,
	}

	if err := svc.initAgent(context.Background()); err != nil {
		return nil, err
	}

	return svc, nil
}
```

- [ ] **Step 3: 在 Chat 方法中集成压缩**

修改 `chat.go` 的 `Chat` 方法，在 `buildMessages` 之后、`LLM.Stream` 之前插入压缩：

```go
// 3. 构建对话消息列表
messages, err := s.buildMessages(ctx, session.ID)
if err != nil {
	_ = cb(&ChatEvent{Error: fmt.Errorf("构建消息失败: %w", err)})
	return
}

// 3.5 上下文压缩（溢出检测 + 自动摘要）
messages, err = s.Compactor.CompactIfNeeded(ctx, messages, session)
if err != nil {
	slog.Error("上下文压缩失败", "error", err)
	// 压缩失败不阻塞对话，继续使用原始消息
}
```

- [ ] **Step 4: 确认 import 语句**

在 `chat.go` 文件头部添加 compactor 包导入（如果启用 goimports 或其他自动管理工具，保存时会自动添加）。

---

### Task 6: 单元测试

**Files:**
- Create: `ai-server/internal/service/compactor/tokenizer_test.go`
- Create: `ai-server/internal/service/compactor/compactor_test.go`

- [ ] **Step 1: tokenizer 测试**

```go
package compactor

import (
	"testing"
)

func TestEstimateTokens_Empty(t *testing.T) {
	if n := EstimateTokens(""); n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}
}

func TestEstimateTokens_Short(t *testing.T) {
	n := EstimateTokens("hello")
	if n == 0 {
		t.Fatal("expected non-zero for short text")
	}
}

func TestEstimateTokens_Long(t *testing.T) {
	text := string(make([]byte, 4000))
	n := EstimateTokens(text)
	if n != 1000 {
		t.Fatalf("expected 1000 for 4000 chars, got %d", n)
	}
}

func TestEstimateMessageTokens(t *testing.T) {
	n := EstimateMessageTokens("hello")
	if n <= 10 {
		t.Fatalf("expected > 10 (content + overhead), got %d", n)
	}
}

func TestEstimateMessagesTokens(t *testing.T) {
	total := EstimateMessagesTokens([]string{"a", "bb", "ccc"})
	if total <= 0 {
		t.Fatal("expected non-zero total")
	}
}
```

- [ ] **Step 2: Run tokenizer tests**

Run: `go test ./ai-server/internal/service/compactor/ -run TestEstimate -v`

Expected: All PASS

- [ ] **Step 3: compactor 单元测试**

```go
package compactor

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"

	"github.com/gangantongxue/knowsync/ai-server/internal/repository"
)

// mockSummaryRepo 模拟摘要存储
type mockSummaryRepo struct {
	summaries map[string]string
}

func (m *mockSummaryRepo) UpdateSummary(sessionID, summary string) error {
	m.summaries[sessionID] = summary
	return nil
}

// mockSummarizer 模拟摘要生成器
type mockSummarizer struct{}

func (m *mockSummarizer) Summarize(ctx context.Context, messages []string, existingSummary string) (string, error) {
	return "这是模拟摘要：用户询问了知识库相关问题。", nil
}

func TestCompactIfNeeded_ShortConversation(t *testing.T) {
	c := New(128_000, &mockSummarizer{}, &mockSummaryRepo{summaries: map[string]string{}})
	msgs := []*schema.Message{
		schema.UserMessage("你好"),
		schema.AssistantMessage("你好！有什么可以帮助你的？", nil),
	}

	result, err := c.CompactIfNeeded(context.Background(), msgs, &repository.ChatSession{ID: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 messages (no compression needed), got %d", len(result))
	}
}

func TestCompactIfNeeded_WithCompression(t *testing.T) {
	c := New(128_000, &mockSummarizer{}, &mockSummaryRepo{summaries: map[string]string{}})

	// 构建超过阈值的消息（大量内容）
	var msgs []*schema.Message
	for i := 0; i < 200; i++ {
		if i%2 == 0 {
			msgs = append(msgs, schema.UserMessage("用户问题 "+string(rune(i))+" "+string(make([]byte, 500))))
		} else {
			msgs = append(msgs, schema.AssistantMessage("助手回答 "+string(rune(i))+" "+string(make([]byte, 500)), nil))
		}
	}

	result, err := c.CompactIfNeeded(context.Background(), msgs, &repository.ChatSession{ID: "test2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 压缩后应有: 1 条摘要 system 消息 + 最近 3 轮(6 条) = 7 条
	if len(result) != 1+KeepRecentTurns*2 {
		t.Fatalf("expected %d messages after compaction, got %d", 1+KeepRecentTurns*2, len(result))
	}

	// 第一条应该是 system 消息
	if result[0].Role != "system" {
		t.Fatalf("first message should be system message (summary), got role: %s", result[0].Role)
	}
}

func TestGroupTurns(t *testing.T) {
	msgs := []*schema.Message{
		schema.UserMessage("q1"),
		schema.AssistantMessage("a1", nil),
		schema.UserMessage("q2"),
		schema.AssistantMessage("a2", nil),
	}

	turns := groupTurns(msgs)
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
}
```

- [ ] **Step 4: Run compactor tests**

Run: `go test ./ai-server/internal/service/compactor/ -v`

Expected: All PASS

---

### Task 7: 编译验证与 lint

- [ ] **Step 1: 编译整个项目**

Run: `go build ./ai-server/...`

Expected: Build succeeds with no errors

- [ ] **Step 2: Run go vet**

Run: `go vet ./ai-server/internal/service/compactor/...`

Expected: No warnings
