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
	if n != 2 {
		t.Fatalf("expected 2 for 'hello' (5 chars / 4), got %d", n)
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
	// 每条消息 1 token + 10 开销 = 11，3 条共 33
	if total != 33 {
		t.Fatalf("expected 33 for 3 short messages, got %d", total)
	}
}
