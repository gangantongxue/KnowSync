package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// SendPrivateMessage 发送私聊消息.
func (s *Service) SendPrivateMessage(ctx context.Context, senderID, receiverID, contentType, content, extra, replyToID string) (*schema.Message, error) {
	// 生成会话ID（排序后拼接）
	conversationID := GetPrivateConversationID(senderID, receiverID)

	// 验证回复消息是否存在且在同一个会话中
	if replyToID != "" {
		if err := s.validateReplyMessage(ctx, ConvTypePrivate, conversationID, replyToID); err != nil {
			return nil, err
		}
	}

	// 获取当前最大seq_id并自增
	maxSeqID, err := s.Repo.Message.GetMaxSeqID(ctx, ConvTypePrivate, conversationID)
	if err != nil {
		slog.Error("获取最大seq_id失败", "error", err)
		return nil, err
	}

	// 创建消息
	now := time.Now().Unix()
	msg := &schema.Message{
		ConversationType: ConvTypePrivate,
		ConversationID:   conversationID,
		SeqID:            maxSeqID + 1,
		SenderID:         senderID,
		ContentType:      contentType,
		Content:          content,
		Status:           "normal",
		CreatedAt:        now,
	}
	if extra != "" {
		msg.Extra = &extra
	}
	if replyToID != "" {
		msg.ReplyToID = &replyToID
	}

	if err := s.Repo.Message.CreateMessage(ctx, msg); err != nil {
		slog.Error("创建消息失败", "error", err)
		return nil, err
	}

	// 更新双方的 last_message_at
	if err := s.Repo.Friend.UpdateFriendLastMessageAt(ctx, senderID, receiverID, now); err != nil {
		slog.Error("更新发送方 last_message_at 失败", "error", err)
	}
	if err := s.Repo.Friend.UpdateFriendLastMessageAt(ctx, receiverID, senderID, now); err != nil {
		slog.Error("更新接收方 last_message_at 失败", "error", err)
	}

	// 推送新消息事件到接收方
	s.pushNewMessageEvent(receiverID, msg)

	return msg, nil
}

// SendGroupMessage 发送群聊消息.
//
//nolint:gocyclo // 群聊消息发送涉及多项校验，保持内聚
func (s *Service) SendGroupMessage(ctx context.Context, senderID, groupID, contentType, content, extra string, mentions []string, replyToID string) (*schema.Message, error) {
	// 检查发送者是否是群成员
	var count int64
	if err := s.Repo.DB.WithContext(ctx).
		Model(&schema.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, senderID).
		Count(&count).Error; err != nil {
		slog.Error("检查群成员失败", "error", err)
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("用户不在群聊中")
	}

	// 验证回复消息是否存在且在同一个群聊中
	if replyToID != "" {
		if err := s.validateReplyMessage(ctx, ConvTypeGroup, groupID, replyToID); err != nil {
			return nil, err
		}
	}

	// 构建Extra JSON（将mentions合并到extra中）
	var extraStr string
	hasExtra := extra != ""
	hasMentions := len(mentions) > 0
	if hasExtra || hasMentions {
		extraData := make(map[string]any)
		if hasExtra {
			if err := json.Unmarshal([]byte(extra), &extraData); err != nil {
				slog.Warn("解析extra失败", "error", err)
			}
		}
		if hasMentions {
			extraData["mentions"] = mentions
		}
		b, _ := json.Marshal(extraData)
		s := string(b)
		extraStr = s
	}

	maxSeqID, err := s.Repo.Message.GetMaxSeqID(ctx, ConvTypeGroup, groupID)
	if err != nil {
		slog.Error("获取最大seq_id失败", "error", err)
		return nil, err
	}

	now := time.Now().Unix()
	msg := &schema.Message{
		ConversationType: ConvTypeGroup,
		ConversationID:   groupID,
		SeqID:            maxSeqID + 1,
		SenderID:         senderID,
		ContentType:      contentType,
		Content:          content,
		Status:           "normal",
		CreatedAt:        now,
	}
	if extraStr != "" {
		msg.Extra = &extraStr
	}
	if replyToID != "" {
		msg.ReplyToID = &replyToID
	}

	if err := s.Repo.Message.CreateMessage(ctx, msg); err != nil {
		slog.Error("创建消息失败", "error", err)
		return nil, err
	}

	// 推送新消息事件到所有群成员（排除发送者）
	s.pushGroupNewMessageEvent(ctx, groupID, senderID, msg)

	return msg, nil
}

// GetMessages 获取消息列表（游标分页）.
func (s *Service) GetMessages(ctx context.Context, conversationType, conversationID string, beforeSeqID uint64, limit int) ([]schema.Message, error) {
	return s.Repo.Message.GetMessagesByConversation(ctx, conversationType, conversationID, beforeSeqID, limit)
}

// RecallMessage 撤回消息.
func (s *Service) RecallMessage(ctx context.Context, messageID, senderID string) error {
	msg, err := s.Repo.Message.GetMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("消息不存在")
		}
		slog.Error("获取消息失败", "error", err)
		return err
	}

	if msg.SenderID != senderID {
		return errors.New("只能撤回自己的消息")
	}

	if time.Now().Unix()-msg.CreatedAt > 300 {
		return errors.New("超过5分钟，无法撤回")
	}

	if err := s.Repo.Message.RecallMessage(ctx, messageID, senderID); err != nil {
		slog.Error("撤回消息失败", "error", err)
		return err
	}

	// 推送消息撤回事件
	s.pushRecallEvent(ctx, msg, senderID)

	return nil
}

// GetUserUnreadCounts 获取用户所有会话的未读数（内部查找用户会话列表）.
func (s *Service) GetUserUnreadCounts(ctx context.Context, userID string) (map[string]int32, error) {
	friends, err := s.Repo.Conversation.GetFriendConversations(ctx, userID)
	if err != nil {
		slog.Error("获取好友会话失败", "error", err)
		return nil, err
	}
	groups, err := s.Repo.Conversation.GetGroupConversations(ctx, userID)
	if err != nil {
		slog.Error("获取群组会话失败", "error", err)
	}

	var conversations []ConversationInfo
	for _, f := range friends {
		conversations = append(conversations, ConversationInfo{
			ConversationType: ConvTypePrivate,
			ConversationID:   GetPrivateConversationID(userID, f.FriendID),
		})
	}
	for _, g := range groups {
		conversations = append(conversations, ConversationInfo{
			ConversationType: ConvTypeGroup,
			ConversationID:   g.ID,
		})
	}

	return s.GetUnreadCounts(ctx, userID, conversations)
}

// GetUnreadCounts 批量获取会话未读数.
func (s *Service) GetUnreadCounts(ctx context.Context, userID string, conversations []ConversationInfo) (map[string]int32, error) {
	result := make(map[string]int32)
	for _, c := range conversations {
		var lastReadSeqID uint64

		switch c.ConversationType {
		case ConvTypePrivate:
			// 从好友关系中查找对方
			parts := strings.Split(c.ConversationID, "_")
			if len(parts) != 2 {
				result[c.ConversationID] = 0
				continue
			}
			friendID := parts[0]
			if friendID == userID {
				friendID = parts[1]
			}
			friend, err := s.Repo.Friend.GetFriend(ctx, userID, friendID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					result[c.ConversationID] = 0
					continue
				}
				slog.Error("获取好友关系失败", "error", err)
				return nil, err
			}
			lastReadSeqID = friend.LastReadSeqID

		case ConvTypeGroup:
			var member schema.GroupMember
			if err := s.Repo.DB.WithContext(ctx).
				Where("group_id = ? AND user_id = ?", c.ConversationID, userID).
				First(&member).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					result[c.ConversationID] = 0
					continue
				}
				slog.Error("获取群成员失败", "error", err)
				return nil, err
			}
			lastReadSeqID = member.LastReadSeqID

		default:
			result[c.ConversationID] = 0
			continue
		}

		count, err := s.Repo.Message.CountMessagesAfter(ctx, c.ConversationType, c.ConversationID, lastReadSeqID)
		if err != nil {
			slog.Error("统计未读消息失败", "error", err)
			return nil, err
		}
		result[c.ConversationID] = int32(count) //nolint:gosec // 未读数不会超过 int32 范围
	}
	return result, nil
}

// GetPrivateConversationID 生成私聊会话ID（排序后拼接）.
func GetPrivateConversationID(id1, id2 string) string {
	id1Int, err1 := strconv.ParseUint(id1, 10, 64)
	id2Int, err2 := strconv.ParseUint(id2, 10, 64)
	if err1 == nil && err2 == nil {
		if id1Int < id2Int {
			return id1 + "_" + id2
		}
		return id2 + "_" + id1
	}
	if strings.Compare(id1, id2) < 0 {
		return id1 + "_" + id2
	}
	return id2 + "_" + id1
}

// validateReplyMessage 验证回复消息是否存在且在同一个会话中.
func (s *Service) validateReplyMessage(ctx context.Context, conversationType, conversationID, replyToID string) error {
	replyMsg, err := s.Repo.Message.GetMessageByID(ctx, replyToID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("回复的消息不存在")
		}
		slog.Error("获取回复消息失败", "error", err)
		return err
	}
	if replyMsg.ConversationType != conversationType || replyMsg.ConversationID != conversationID {
		return errors.New("回复的消息不在当前会话中")
	}
	return nil
}

// ForwardMessage 转发消息.
//
//nolint:gocyclo // 转发逻辑涉及多步骤处理，保持内聚
func (s *Service) ForwardMessage(ctx context.Context, senderID, targetConversationType, targetConversationID string, messageIDs []string) (*schema.Message, error) {
	// 验证发送者是否在目标会话中
	switch targetConversationType {
	case ConvTypePrivate:
		parts := strings.Split(targetConversationID, "_")
		if len(parts) != 2 {
			return nil, errors.New("无效的私聊会话ID")
		}
		var otherUserID string
		switch {
		case parts[0] == senderID:
			otherUserID = parts[1]
		case parts[1] == senderID:
			otherUserID = parts[0]
		default:
			return nil, errors.New("用户不在该私聊会话中")
		}
		exists, err := s.Repo.Friend.CheckFriendExists(ctx, senderID, otherUserID)
		if err != nil {
			slog.Error("检查好友关系失败", "error", err)
			return nil, err
		}
		if !exists {
			return nil, errors.New("用户不在该私聊会话中")
		}
	case ConvTypeGroup:
		var count int64
		if err := s.Repo.DB.WithContext(ctx).
			Model(&schema.GroupMember{}).
			Where("group_id = ? AND user_id = ?", targetConversationID, senderID).
			Count(&count).Error; err != nil {
			slog.Error("检查群成员失败", "error", err)
			return nil, err
		}
		if count == 0 {
			return nil, errors.New("用户不在群聊中")
		}
	default:
		return nil, errors.New("无效的会话类型")
	}

	// 获取原始消息
	originalMessages, err := s.Repo.Message.GetMessagesByIDs(ctx, messageIDs)
	if err != nil {
		slog.Error("获取原始消息失败", "error", err)
		return nil, err
	}
	if len(originalMessages) == 0 {
		return nil, errors.New("消息不存在")
	}

	// 获取原始消息发送者名称
	senderIDs := make([]string, 0)
	senderIDSet := make(map[string]struct{})
	for _, msg := range originalMessages {
		if _, ok := senderIDSet[msg.SenderID]; !ok {
			senderIDSet[msg.SenderID] = struct{}{}
			senderIDs = append(senderIDs, msg.SenderID)
		}
	}
	senderNames, err := s.getUserNames(ctx, senderIDs)
	if err != nil {
		slog.Error("获取用户名称失败", "error", err)
		return nil, err
	}

	// 拼接转发内容
	var contentBuilder strings.Builder
	for i, msg := range originalMessages {
		if i > 0 {
			contentBuilder.WriteString("\n\n")
		}
		name := senderNames[msg.SenderID]
		if name == "" {
			name = msg.SenderID
		}
		fmt.Fprintf(&contentBuilder, "「%s」\n\n——来自 %s", msg.Content, name)
	}

	// 构建Extra，记录转发信息和原始内容类型
	extraData := map[string]any{
		"forward":              true,
		"original_message_ids": messageIDs,
	}
	if len(originalMessages) == 1 {
		extraData["original_content_type"] = originalMessages[0].ContentType
	}
	extraBytes, _ := json.Marshal(extraData)
	extraStr := string(extraBytes)

	// 创建转发消息
	maxSeqID, err := s.Repo.Message.GetMaxSeqID(ctx, targetConversationType, targetConversationID)
	if err != nil {
		slog.Error("获取最大seq_id失败", "error", err)
		return nil, err
	}

	now := time.Now().Unix()
	msg := &schema.Message{
		ConversationType: targetConversationType,
		ConversationID:   targetConversationID,
		SeqID:            maxSeqID + 1,
		SenderID:         senderID,
		ContentType:      "text",
		Content:          contentBuilder.String(),
		Extra:            &extraStr,
		Status:           "normal",
		CreatedAt:        now,
	}

	if err := s.Repo.Message.CreateMessage(ctx, msg); err != nil {
		slog.Error("创建转发消息失败", "error", err)
		return nil, err
	}

	// 更新私聊会话的 last_message_at
	if targetConversationType == ConvTypePrivate {
		s.updatePrivateConversationLastMessageAt(ctx, senderID, targetConversationID, now)
	}

	return msg, nil
}

// getUserNames 批量获取用户名称.
func (s *Service) getUserNames(ctx context.Context, userIDs []string) (map[string]string, error) {
	var users []schema.User
	if err := s.Repo.DB.WithContext(ctx).
		Where("id IN ?", userIDs).
		Find(&users).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(users))
	for _, u := range users {
		result[u.ID] = u.Name
	}
	return result, nil
}

// pushNewMessageEvent 向指定用户推送新消息事件.
func (s *Service) pushNewMessageEvent(userID string, msg *schema.Message) {
	data := map[string]any{
		"id":                msg.ID,
		"conversation_type": msg.ConversationType,
		"conversation_id":   msg.ConversationID,
		"seq_id":            msg.SeqID,
		"sender_id":         msg.SenderID,
		"content_type":      msg.ContentType,
		"content":           msg.Content,
		"extra":             msg.Extra,
		"reply_to_id":       msg.ReplyToID,
		"status":            msg.Status,
		"created_at":        msg.CreatedAt,
	}
	payload, err := json.Marshal(map[string]any{
		PushKeyType: PushTypeNewMessage,
		PushKeyData: data,
	})
	if err != nil {
		slog.Error("序列化新消息事件失败", "error", err)
		return
	}
	s.Hub.SendToUser(userID, payload)
}

// pushGroupNewMessageEvent 向群成员推送新消息事件（排除发送者）.
func (s *Service) pushGroupNewMessageEvent(ctx context.Context, groupID, senderID string, msg *schema.Message) {
	var memberIDs []string
	if err := s.Repo.DB.WithContext(ctx).
		Model(&schema.GroupMember{}).
		Where("group_id = ?", groupID).
		Pluck("user_id", &memberIDs).Error; err != nil {
		slog.Error("获取群成员列表失败", "error", err)
		return
	}
	for _, memberID := range memberIDs {
		if memberID == senderID {
			continue
		}
		s.pushNewMessageEvent(memberID, msg)
	}
}

// pushRecallEvent 推送消息撤回事件.
func (s *Service) pushRecallEvent(ctx context.Context, msg *schema.Message, senderID string) {
	switch msg.ConversationType {
	case ConvTypePrivate:
		// 私聊：推送给对方
		parts := strings.Split(msg.ConversationID, "_")
		if len(parts) != 2 {
			return
		}
		var targetID string
		if parts[0] == senderID {
			targetID = parts[1]
		} else {
			targetID = parts[0]
		}
		s.pushRecallToUser(targetID, msg.ID, msg.ConversationType, msg.ConversationID)
	case ConvTypeGroup:
		// 群聊：推送给所有成员（排除发送者）
		var memberIDs []string
		if err := s.Repo.DB.WithContext(ctx).
			Model(&schema.GroupMember{}).
			Where("group_id = ?", msg.ConversationID).
			Pluck("user_id", &memberIDs).Error; err != nil {
			slog.Error("获取群成员列表失败", "error", err)
			return
		}
		for _, memberID := range memberIDs {
			if memberID == senderID {
				continue
			}
			s.pushRecallToUser(memberID, msg.ID, msg.ConversationType, msg.ConversationID)
		}
	}
}

// pushRecallToUser 向指定用户发送撤回事件.
func (s *Service) pushRecallToUser(userID, messageID, conversationType, conversationID string) {
	payload, err := json.Marshal(map[string]any{
		"type": "message_recalled",
		"data": map[string]any{
			"message_id":        messageID,
			"conversation_type": conversationType,
			"conversation_id":   conversationID,
		},
	})
	if err != nil {
		slog.Error("序列化撤回事件失败", "error", err)
		return
	}
	s.Hub.SendToUser(userID, payload)
}

// updatePrivateConversationLastMessageAt 更新私聊会话双方的 last_message_at.
func (s *Service) updatePrivateConversationLastMessageAt(ctx context.Context, senderID, targetConversationID string, now int64) {
	parts := strings.Split(targetConversationID, "_")
	if len(parts) != 2 {
		return
	}
	var receiverID string
	if parts[0] == senderID {
		receiverID = parts[1]
	} else {
		receiverID = parts[0]
	}
	if err := s.Repo.Friend.UpdateFriendLastMessageAt(ctx, senderID, receiverID, now); err != nil {
		slog.Error("更新发送方 last_message_at 失败", "error", err)
	}
	if err := s.Repo.Friend.UpdateFriendLastMessageAt(ctx, receiverID, senderID, now); err != nil {
		slog.Error("更新接收方 last_message_at 失败", "error", err)
	}
}

// ConversationInfo 会话信息.
type ConversationInfo struct {
	ConversationType string
	ConversationID   string
}
