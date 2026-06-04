package compactor

import (
	"math"
)

const avgCharPerToken = 4

// EstimateTokens 估算文本的 token 数（4 字符 ≈ 1 token）.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return int(math.Ceil(float64(len(text)) / avgCharPerToken))
}

// EstimateMessageTokens 估算单条消息的 token 数（消息体 + 角色/格式开销）.
func EstimateMessageTokens(content string) int {
	return EstimateTokens(content) + 10 // 10 token 角色/格式开销
}

// EstimateMessagesTokens 估算多条消息的总 token 数.
func EstimateMessagesTokens(contents []string) int {
	total := 0
	for _, c := range contents {
		total += EstimateMessageTokens(c)
	}
	return total
}
