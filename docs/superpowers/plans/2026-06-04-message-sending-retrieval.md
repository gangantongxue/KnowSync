# Task 2.2: 消息发送与获取 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement message sending and retrieval functionality in chat-server

**Architecture:** Create MessageRepository and MessageService following existing patterns, update proto definitions with message-related RPCs

**Tech Stack:** Go, GORM, gRPC, Protocol Buffers

---

## Task 1: Update Proto Definitions

**Files:**
- Modify: `ks-proto/proto/chat.proto`

- [ ] **Step 1: Add message-related message definitions**

Add to chat.proto after the SearchUserInfo message:

```protobuf
// Message 消息
message Message {
  string id = 1;
  string conversation_type = 2;
  string conversation_id = 3;
  uint64 seq_id = 4;
  string sender_id = 5;
  string content_type = 6;
  string content = 7;
  string extra = 8;
  string reply_to_id = 9;
  string status = 10;
  int64 created_at = 11;
}

// SendPrivateMessageReq 发送私聊消息请求
message SendPrivateMessageReq {
  string sender_id = 1;
  string receiver_id = 2;
  string content_type = 3;
  string content = 4;
  string extra = 5;
  string reply_to_id = 6;
}

// SendPrivateMessageResp 发送私聊消息响应
message SendPrivateMessageResp {
  bool success = 1;
  string msg = 2;
  Message message = 3;
}

// SendGroupMessageReq 发送群聊消息请求
message SendGroupMessageReq {
  string sender_id = 1;
  string group_id = 2;
  string content_type = 3;
  string content = 4;
  string extra = 5;
  repeated string mentions = 6;
  string reply_to_id = 7;
}

// SendGroupMessageResp 发送群聊消息响应
message SendGroupMessageResp {
  bool success = 1;
  string msg = 2;
  Message message = 3;
}

// GetMessagesReq 获取消息列表请求
message GetMessagesReq {
  string conversation_type = 1;
  string conversation_id = 2;
  uint64 before_seq_id = 3;
  int32 limit = 4;
}

// GetMessagesResp 获取消息列表响应
message GetMessagesResp {
  bool success = 1;
  string msg = 2;
  repeated Message messages = 3;
}

// RecallMessageReq 撤回消息请求
message RecallMessageReq {
  string message_id = 1;
  string sender_id = 2;
}

// RecallMessageResp 撤回消息响应
message RecallMessageResp {
  bool success = 1;
  string msg = 2;
}

// GetUnreadCountReq 获取未读数请求
message GetUnreadCountReq {
  string user_id = 1;
  repeated Conversation conversation = 2;
}

// Conversation 会话信息
message Conversation {
  string conversation_type = 1;
  string conversation_id = 2;
  uint64 last_read_seq_id = 3;
}

// GetUnreadCountResp 获取未读数响应
message GetUnreadCountResp {
  bool success = 1;
  string msg = 2;
  map<string, int32> counts = 3;
}
```

- [ ] **Step 2: Add RPCs to ChatService**

Add to ChatService definition:

```protobuf
// SendPrivateMessage 发送私聊消息
rpc SendPrivateMessage(SendPrivateMessageReq) returns (SendPrivateMessageResp);
// SendGroupMessage 发送群聊消息
rpc SendGroupMessage(SendGroupMessageReq) returns (SendGroupMessageResp);
// GetMessages 获取消息列表
rpc GetMessages(GetMessagesReq) returns (GetMessagesResp);
// RecallMessage 撤回消息
rpc RecallMessage(RecallMessageReq) returns (RecallMessageResp);
// GetUnreadCount 获取未读数
rpc GetUnreadCount(GetUnreadCountReq) returns (GetUnreadCountResp);
```

- [ ] **Step 3: Generate protobuf code**

Run: `cd ks-proto && buf generate` or `protoc --go_out=. --go-grpc_out=. proto/chat.proto`

- [ ] **Step 4: Verify build passes**

Run: `go build ./ks-proto/...`

---

## Task 2: Create MessageRepository

**Files:**
- Create: `chat-server/internal/repository/message.go`

- [ ] **Step 1: Create MessageRepository struct and constructor**

```go
package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// MessageRepository 消息仓库
type MessageRepository struct {
	DB *gorm.DB
}

// NewMessageRepository 创建消息仓库
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}
```

- [ ] **Step 2: Implement CreateMessage**

```go
// CreateMessage 创建消息
func (r *MessageRepository) CreateMessage(ctx context.Context, msg *schema.Message) error {
	return r.DB.WithContext(ctx).Create(msg).Error
}
```

- [ ] **Step 3: Implement GetMessagesByConversation**

```go
// GetMessagesByConversation 根据会话获取消息列表
// beforeSeqID 用于游标分页，返回 seq_id < beforeSeqID 的消息
// limit 限制返回数量，默认50，最大100
func (r *MessageRepository) GetMessagesByConversation(ctx context.Context, conversationType, conversationID string, beforeSeqID uint64, limit int) ([]schema.Message, error) {
	var messages []schema.Message

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := r.DB.WithContext(ctx).
		Where("conversation_type = ? AND conversation_id = ?", conversationType, conversationID)

	if beforeSeqID > 0 {
		query = query.Where("seq_id < ?", beforeSeqID)
	}

	if err := query.Order("seq_id DESC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}
```

- [ ] **Step 4: Implement GetMessageByID**

```go
// GetMessageByID 根据ID获取消息
func (r *MessageRepository) GetMessageByID(ctx context.Context, messageID string) (*schema.Message, error) {
	var msg schema.Message
	if err := r.DB.WithContext(ctx).Where("id = ?", messageID).First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}
```

- [ ] **Step 5: Implement GetLastMessageByConversation**

```go
// GetLastMessageByConversation 获取会话最新消息
func (r *MessageRepository) GetLastMessageByConversation(ctx context.Context, conversationType, conversationID string) (*schema.Message, error) {
	var msg schema.Message
	if err := r.DB.WithContext(ctx).
		Where("conversation_type = ? AND conversation_id = ?", conversationType, conversationID).
		Order("seq_id DESC").
		First(&msg).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}
```

- [ ] **Step 6: Implement GetUnreadCount**

```go
// GetUnreadCount 获取未读消息数
// 统计 seq_id > lastReadSeqID 的消息数量
func (r *MessageRepository) GetUnreadCount(ctx context.Context, conversationType, conversationID string, lastReadSeqID uint64) (int64, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&schema.Message{}).
		Where("conversation_type = ? AND conversation_id = ? AND seq_id > ?", conversationType, conversationID, lastReadSeqID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
```

- [ ] **Step 7: Implement RecallMessage**

```go
// RecallMessage 撤回消息
func (r *MessageRepository) RecallMessage(ctx context.Context, messageID, senderID string) error {
	return r.DB.WithContext(ctx).
		Model(&schema.Message{}).
		Where("id = ? AND sender_id = ?", messageID, senderID).
		Update("status", "recalled").Error
}
```

- [ ] **Step 8: Implement GetMessagesByIDs**

```go
// GetMessagesByIDs 批量获取消息
func (r *MessageRepository) GetMessagesByIDs(ctx context.Context, messageIDs []string) ([]schema.Message, error) {
	var messages []schema.Message
	if err := r.DB.WithContext(ctx).Where("id IN ?", messageIDs).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}
```

- [ ] **Step 9: Verify build passes**

Run: `go build ./chat-server/...`

---

## Task 3: Create MessageService

**Files:**
- Create: `chat-server/internal/service/message.go`

- [ ] **Step 1: Create MessageService struct and constructor**

```go
package service

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/gangantongxue/knowsync/chat-server/internal/repository"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// MessageService 消息服务
type MessageService struct {
	messageRepo *repository.MessageRepository
	friendRepo  *repository.FriendRepository
	groupRepo   *repository.GroupRepository
}

// NewMessageService 创建消息服务
func NewMessageService(messageRepo *repository.MessageRepository, friendRepo *repository.FriendRepository, groupRepo *repository.GroupRepository) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		friendRepo:  friendRepo,
		groupRepo:   groupRepo,
	}
}
```

- [ ] **Step 2: Implement SendPrivateMessage**

```go
// SendPrivateMessage 发送私聊消息
func (s *MessageService) SendPrivateMessage(ctx context.Context, senderID, receiverID, contentType, content, extra string, replyToID string) (*schema.Message, error) {
	// 生成会话ID：将两个用户ID排序后拼接
	userIDs := []string{senderID, receiverID}
	sort.Strings(userIDs)
	conversationID := strings.Join(userIDs, "_")

	// 获取当前会话的最大seq_id
	lastMsg, err := s.messageRepo.GetLastMessageByConversation(ctx, "private", conversationID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("获取最新消息失败", "error", err)
		return nil, err
	}

	var seqID uint64 = 1
	if lastMsg != nil {
		seqID = lastMsg.SeqID + 1
	}

	// 创建消息
	now := time.Now().Unix()
	msg := &schema.Message{
		ConversationType: "private",
		ConversationID:   conversationID,
		SeqID:            seqID,
		SenderID:         senderID,
		ContentType:      contentType,
		Content:          content,
		Extra:            &extra,
		Status:           "normal",
		CreatedAt:        now,
	}

	if replyToID != "" {
		msg.ReplyToID = &replyToID
	}

	if err := s.messageRepo.CreateMessage(ctx, msg); err != nil {
		slog.Error("创建消息失败", "error", err)
		return nil, err
	}

	// 更新好友关系的最后消息时间
	now = time.Now().Unix()
	if err := s.friendRepo.UpdateFriendLastMessageAt(ctx, senderID, receiverID, now); err != nil {
		slog.Error("更新好友最后消息时间失败", "error", err)
	}

	return msg, nil
}
```

- [ ] **Step 3: Implement SendGroupMessage**

```go
// SendGroupMessage 发送群聊消息
func (s *MessageService) SendGroupMessage(ctx context.Context, senderID, groupID, contentType, content, extra string, mentions []string, replyToID string) (*schema.Message, error) {
	// 获取当前会话的最大seq_id
	lastMsg, err := s.messageRepo.GetLastMessageByConversation(ctx, "group", groupID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("获取最新消息失败", "error", err)
		return nil, err
	}

	var seqID uint64 = 1
	if lastMsg != nil {
		seqID = lastMsg.SeqID + 1
	}

	// 处理mentions信息
	var extraPtr *string
	if extra != "" || len(mentions) > 0 {
		extraStr := extra
		if len(mentions) > 0 {
			// 将mentions添加到extra JSON中
			if extraStr == "" {
				extraStr = "{}"
			}
			extraStr = strings.TrimSuffix(extraStr, "}")
			if extraStr != "{" {
				extraStr += ","
			}
			extraStr += `"mentions":["` + strings.Join(mentions, `","`) + `"]}`
		}
		extraPtr = &extraStr
	}

	// 创建消息
	now := time.Now().Unix()
	msg := &schema.Message{
		ConversationType: "group",
		ConversationID:   groupID,
		SeqID:            seqID,
		SenderID:         senderID,
		ContentType:      contentType,
		Content:          content,
		Extra:            extraPtr,
		Status:           "normal",
		CreatedAt:        now,
	}

	if replyToID != "" {
		msg.ReplyToID = &replyToID
	}

	if err := s.messageRepo.CreateMessage(ctx, msg); err != nil {
		slog.Error("创建消息失败", "error", err)
		return nil, err
	}

	// 更新群组成员的最后阅读时间（除发送者外）
	if err := s.groupRepo.UpdateGroupMembersLastReadAt(ctx, groupID, senderID, now); err != nil {
		slog.Error("更新群组成员最后阅读时间失败", "error", err)
	}

	return msg, nil
}
```

- [ ] **Step 4: Implement GetMessages**

```go
// GetMessages 获取消息列表
func (s *MessageService) GetMessages(ctx context.Context, conversationType, conversationID string, beforeSeqID uint64, limit int) ([]schema.Message, error) {
	return s.messageRepo.GetMessagesByConversation(ctx, conversationType, conversationID, beforeSeqID, limit)
}
```

- [ ] **Step 5: Implement RecallMessage**

```go
// RecallMessage 撤回消息
// 只有发送者可以在5分钟内撤回消息
func (s *MessageService) RecallMessage(ctx context.Context, messageID, senderID string) error {
	// 获取消息
	msg, err := s.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("消息不存在")
		}
		slog.Error("获取消息失败", "error", err)
		return err
	}

	// 验证发送者
	if msg.SenderID != senderID {
		return errors.New("只能撤回自己发送的消息")
	}

	// 验证是否超过5分钟
	if time.Now().Unix()-msg.CreatedAt > 5*60 {
		return errors.New("消息发送超过5分钟，无法撤回")
	}

	// 验证消息状态
	if msg.Status == "recalled" {
		return errors.New("消息已被撤回")
	}

	// 撤回消息
	if err := s.messageRepo.RecallMessage(ctx, messageID, senderID); err != nil {
		slog.Error("撤回消息失败", "error", err)
		return err
	}

	return nil
}
```

- [ ] **Step 6: Implement GetUnreadCounts**

```go
// GetUnreadCounts 批量获取未读消息数
func (s *MessageService) GetUnreadCounts(ctx context.Context, userID string, conversations []*schema.Conversation) (map[string]int32, error) {
	counts := make(map[string]int32)

	for _, conv := range conversations {
		count, err := s.messageRepo.GetUnreadCount(ctx, conv.ConversationType, conv.ConversationID, conv.LastReadSeqID)
		if err != nil {
			slog.Error("获取未读数失败", "conversation", conv.ConversationID, "error", err)
			return nil, err
		}
		// 使用 conversationType:conversationID 作为key
		key := conv.ConversationType + ":" + conv.ConversationID
		counts[key] = int32(count)
	}

	return counts, nil
}
```

- [ ] **Step 7: Verify build passes**

Run: `go build ./chat-server/...`

---

## Task 4: Add Missing Repository Methods

**Files:**
- Modify: `chat-server/internal/repository/friend.go`
- Create: `chat-server/internal/repository/group.go`

- [ ] **Step 1: Add UpdateFriendLastMessageAt to FriendRepository**

Add to friend.go:

```go
// UpdateFriendLastMessageAt 更新好友关系的最后消息时间
func (r *FriendRepository) UpdateFriendLastMessageAt(ctx context.Context, userID, friendID string, lastMessageAt int64) error {
	return r.DB.WithContext(ctx).
		Model(&schema.Friend{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userID, friendID, friendID, userID).
		Update("last_message_at", lastMessageAt).Error
}
```

- [ ] **Step 2: Create GroupRepository**

Create chat-server/internal/repository/group.go:

```go
package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// GroupRepository 群组仓库
type GroupRepository struct {
	DB *gorm.DB
}

// NewGroupRepository 创建群组仓库
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{DB: db}
}

// UpdateGroupMembersLastReadAt 更新群组成员最后阅读时间
// 更新除senderID外的所有成员
func (r *GroupRepository) UpdateGroupMembersLastReadAt(ctx context.Context, groupID, senderID string, lastReadAt int64) error {
	return r.DB.WithContext(ctx).
		Model(&schema.GroupMember{}).
		Where("group_id = ? AND user_id != ?", groupID, senderID).
		Update("last_read_at", lastReadAt).Error
}

// GetGroupMembers 获取群组成员列表
func (r *GroupRepository) GetGroupMembers(ctx context.Context, groupID string) ([]schema.GroupMember, error) {
	var members []schema.GroupMember
	if err := r.DB.WithContext(ctx).
		Where("group_id = ?", groupID).
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}
```

- [ ] **Step 3: Verify build passes**

Run: `go build ./chat-server/...`

---

## Task 5: Register Dependencies in App

**Files:**
- Modify: `chat-server/internal/app/app.go`

- [ ] **Step 1: Update app.go to register new repositories and services**

Update the initialization section in app.go:

```go
// 初始化仓库层和业务层
friendRepo := repository.NewFriendRepository(db.DB)
friendService := service.NewFriendService(friendRepo)

messageRepo := repository.NewMessageRepository(db.DB)
groupRepo := repository.NewGroupRepository(db.DB)
messageService := service.NewMessageService(messageRepo, friendRepo, groupRepo)

_ = friendService   // 供后续 gRPC 服务使用
_ = messageService  // 供后续 gRPC 服务使用
```

- [ ] **Step 2: Verify build passes**

Run: `go build ./chat-server/...`

---

## Task 6: Final Verification

- [ ] **Step 1: Verify all builds pass**

Run: `go build ./chat-server/... && go build ./ks-proto/...`

- [ ] **Step 2: Run any existing tests**

Run: `go test ./chat-server/... ./ks-proto/...`
