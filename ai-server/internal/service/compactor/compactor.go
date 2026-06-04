// Package compactor 提供上下文压缩功能，通过摘要压缩历史消息.
package compactor

import (
	"context"
	"log/slog"

	"github.com/cloudwego/eino/schema"

	"github.com/gangantongxue/knowsync/ai-server/internal/repository"
)

const (
	// OverflowRatio 溢出检测.
	OverflowRatio = 0.8
	// ReservedTokens 保留 token 数，用于避免超出上下文窗口.
	ReservedTokens = 20_000

	// KeepRecentTurns 摘要压缩.
	KeepRecentTurns = 3
	// MaxSummaryLen 摘要最大长度.
	MaxSummaryLen = 2_000
)

// SummaryRepository 摘要存储接口.
type SummaryRepository interface {
	UpdateSummary(sessionID, summary string) error
}

// SummarizerInterface 摘要生成器接口，便于单元测试.
type SummarizerInterface interface {
	Summarize(ctx context.Context, messages []string, existingSummary string) (string, error)
}

// Compactor 上下文管理器.
type Compactor struct {
	llmContextLimit int
	summarizer      SummarizerInterface
	repo            SummaryRepository
}

// New 创建上下文管理器.
func New(llmContextLimit int, summarizer SummarizerInterface, repo SummaryRepository) *Compactor {
	return &Compactor{
		llmContextLimit: llmContextLimit,
		summarizer:      summarizer,
		repo:            repo,
	}
}

// CompactIfNeeded 检测是否需要压缩，需要则执行.
func (c *Compactor) CompactIfNeeded(ctx context.Context, messages []*schema.Message, session *repository.ChatSession) ([]*schema.Message, error) {
	if len(messages) < 3 {
		return messages, nil
	}

	threshold := int(float64(c.llmContextLimit-ReservedTokens) * OverflowRatio)
	totalTokens := estimateMessagesTokenCount(messages)

	if totalTokens < threshold {
		return messages, nil
	}

	slog.Info(
		"上下文即将溢出，开始压缩",
		"session_id", session.ID,
		"total_tokens", totalTokens,
		"threshold", threshold,
	)

	return c.compact(ctx, messages, session)
}

// compact 执行压缩：保留最近 N 轮，将早期消息摘要.
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
		summary = string([]rune(summary)[:MaxSummaryLen])
	}

	// 持久化摘要
	if err := c.repo.UpdateSummary(session.ID, summary); err != nil {
		slog.Error("保存摘要失败", "error", err)
	}

	slog.Info(
		"上下文压缩完成",
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

// estimateMessagesTokenCount 估算消息列表的总 token 数.
func estimateMessagesTokenCount(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		total += EstimateMessageTokens(msg.Content)
	}
	return total
}

// groupTurns 将消息按"用户+助手"轮次分组
// 每条 user 消息开始新的一轮，后续的 assistant 消息归入同一轮.
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

// flattenTurns 将分组消息展平.
func flattenTurns(turns [][]*schema.Message) []*schema.Message {
	var result []*schema.Message
	for _, t := range turns {
		result = append(result, t...)
	}
	return result
}
