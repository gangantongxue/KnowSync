package service

import (
	"encoding/json"
	"sort"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// Conversation 会话统一视图.
type Conversation struct {
	ConversationType string
	ConversationID   string
	Name             string
	Avatar           string
	LastMessage      *schema.Message
	UnreadCount      int32
	LastMessageAt    int64
	Pinned           bool
	Mentioned        bool
}

// ConversationInfo 会话信息.
type ConversationInfo struct {
	ConversationType string
	ConversationID   string
}

// checkMentioned 检查消息是否 @ 了指定用户.
func checkMentioned(msg *schema.Message, userID string) bool {
	if msg == nil || msg.Extra == nil {
		return false
	}
	var extraData map[string]any
	if err := json.Unmarshal([]byte(*msg.Extra), &extraData); err != nil {
		return false
	}
	mentions, ok := extraData["mentions"]
	if !ok {
		return false
	}
	mentionList, ok := mentions.([]any)
	if !ok {
		return false
	}
	for _, m := range mentionList {
		if id, ok := m.(string); ok && id == userID {
			return true
		}
	}
	return false
}

// sortConversations 排序会话：置顶优先，然后按 last_message_at DESC.
func sortConversations(conversations []Conversation) {
	sort.Slice(conversations, func(i, j int) bool {
		if conversations[i].Pinned != conversations[j].Pinned {
			return conversations[i].Pinned
		}
		return conversations[i].LastMessageAt > conversations[j].LastMessageAt
	})
}
