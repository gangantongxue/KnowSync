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
	repo := &mockSummaryRepo{summaries: map[string]string{}}
	c := New(128_000, &mockSummarizer{}, repo)

	// 构建超过阈值的消息（大量内容）
	var msgs []*schema.Message
	for i := 0; i < 200; i++ {
		if i%2 == 0 {
			msgs = append(msgs, schema.UserMessage("用户问题 "+string(rune(i))+" "+string(make([]byte, 4000))))
		} else {
			msgs = append(msgs, schema.AssistantMessage("助手回答 "+string(rune(i))+" "+string(make([]byte, 4000)), nil))
		}
	}

	result, err := c.CompactIfNeeded(context.Background(), msgs, &repository.ChatSession{ID: "test2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 压缩后应有: 1 条摘要 system 消息 + 最近 3 轮(6 条) = 7 条
	expectedLen := 1 + KeepRecentTurns*2
	if len(result) != expectedLen {
		t.Fatalf("expected %d messages after compaction, got %d", expectedLen, len(result))
	}

	// 第一条应该是 system 消息
	if result[0].Role != "system" {
		t.Fatalf("first message should be system message (summary), got role: %s", result[0].Role)
	}

	// 验证摘要已持久化到 mock repo
	sessionID := "test2"
	if _, ok := repo.summaries[sessionID]; !ok {
		t.Fatalf("expected summary to be persisted for session %s", sessionID)
	}
	if repo.summaries[sessionID] == "" {
		t.Fatal("expected non-empty summary")
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
