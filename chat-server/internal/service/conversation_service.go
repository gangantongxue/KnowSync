package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/gangantongxue/knowsync/chat-server/internal/repository"
	"github.com/gangantongxue/knowsync/chat-server/internal/ws"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// ConversationService 会话业务服务.
type ConversationService struct {
	Repo *repository.Repository
	Hub  *ws.Hub
}

// NewConversationService 创建会话服务.
func NewConversationService(repo *repository.Repository, hub *ws.Hub) *ConversationService {
	return &ConversationService{Repo: repo, Hub: hub}
}

// GetConversationList 获取会话列表.
func (s *ConversationService) GetConversationList(ctx context.Context, userID string) ([]Conversation, error) { //nolint:gocyclo // 会话列表聚合逻辑复杂
	friends, err := s.Repo.Conversation.GetFriendConversations(ctx, userID)
	if err != nil {
		slog.Error("获取好友会话失败", "error", err)
		return nil, err
	}

	groups, err := s.Repo.Conversation.GetGroupConversations(ctx, userID)
	if err != nil {
		slog.Error("获取群组会话失败", "error", err)
	}

	var conversations []Conversation

	for _, f := range friends {
		conversationID := GetPrivateConversationID(userID, f.FriendID)

		name := f.FriendID
		if f.Remark != "" {
			name = f.Remark
		}

		var lastMsg *schema.Message
		msg, err := s.Repo.Message.GetLastMessageByConversation(ctx, ConvTypePrivate, conversationID)
		if err == nil {
			lastMsg = msg
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("获取私聊最后消息失败", "error", err, "conversation_id", conversationID)
		}

		unreadCount := int32(0)
		if f.LastReadSeqID > 0 {
			count, err := s.Repo.Message.CountMessagesAfter(ctx, ConvTypePrivate, conversationID, f.LastReadSeqID)
			if err == nil {
				unreadCount = int32(count) //nolint:gosec // 未读数不会超过 int32 范围
			} else {
				slog.Error("统计私聊未读消息失败", "error", err, "conversation_id", conversationID)
			}
		} else {
			count, err := s.Repo.Message.CountMessagesAfter(ctx, ConvTypePrivate, conversationID, 0)
			if err == nil {
				unreadCount = int32(count) //nolint:gosec // 未读数不会超过 int32 范围
			}
		}

		conversations = append(conversations, Conversation{
			ConversationType: ConvTypePrivate,
			ConversationID:   conversationID,
			Name:             name,
			Avatar:           "",
			LastMessage:      lastMsg,
			UnreadCount:      unreadCount,
			LastMessageAt:    f.LastMessageAt,
			Pinned:           f.Pinned == 1,
			Mentioned:        false,
		})
	}

	for _, g := range groups {
		lastMsg, err := s.Repo.Message.GetLastMessageByConversation(ctx, ConvTypeGroup, g.ID)
		var lastMsgPtr *schema.Message
		if err == nil {
			lastMsgPtr = lastMsg
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("获取群聊最后消息失败", "error", err, "group_id", g.ID)
		}

		mentioned := checkMentioned(lastMsgPtr, userID)

		unreadCount := int32(0)
		if g.LastReadSeqID > 0 {
			count, err := s.Repo.Message.CountMessagesAfter(ctx, ConvTypeGroup, g.ID, g.LastReadSeqID)
			if err == nil {
				unreadCount = int32(count) //nolint:gosec // 未读数不会超过 int32 范围
			} else {
				slog.Error("统计群聊未读消息失败", "error", err, "group_id", g.ID)
			}
		} else {
			count, err := s.Repo.Message.CountMessagesAfter(ctx, ConvTypeGroup, g.ID, 0)
			if err == nil {
				unreadCount = int32(count) //nolint:gosec // 未读数不会超过 int32 范围
			}
		}

		lastMessageAt := int64(0)
		if lastMsgPtr != nil {
			lastMessageAt = lastMsgPtr.CreatedAt
		}

		conversations = append(conversations, Conversation{
			ConversationType: ConvTypeGroup,
			ConversationID:   g.ID,
			Name:             g.Name,
			Avatar:           g.Avatar,
			LastMessage:      lastMsgPtr,
			UnreadCount:      unreadCount,
			LastMessageAt:    lastMessageAt,
			Pinned:           g.Pinned == 1,
			Mentioned:        mentioned,
		})
	}

	sortConversations(conversations)

	return conversations, nil
}

// MarkConversationRead 标记会话已读.
func (s *ConversationService) MarkConversationRead(ctx context.Context, userID, conversationType, conversationID string) error {
	maxSeqID, err := s.Repo.Message.GetMaxSeqID(ctx, conversationType, conversationID)
	if err != nil {
		slog.Error("获取最大 seq_id 失败", "error", err)
		return err
	}

	switch conversationType {
	case ConvTypePrivate:
		parts := strings.Split(conversationID, "_")
		if len(parts) != 2 {
			return errors.New("无效的会话ID")
		}
		friendID := parts[0]
		if friendID == userID {
			friendID = parts[1]
		}
		if err := s.Repo.Conversation.UpdateFriendLastReadSeqID(ctx, userID, friendID, maxSeqID); err != nil {
			slog.Error("更新好友已读 seq_id 失败", "error", err)
			return err
		}
	case ConvTypeGroup:
		if err := s.Repo.Conversation.UpdateGroupMemberLastReadSeqID(ctx, conversationID, userID, maxSeqID); err != nil {
			slog.Error("更新群成员已读 seq_id 失败", "error", err)
			return err
		}
	default:
		return errors.New("无效的会话类型")
	}

	return nil
}

// TogglePin 切换置顶.
func (s *ConversationService) TogglePin(ctx context.Context, userID, conversationType, conversationID string) (bool, error) {
	switch conversationType {
	case ConvTypePrivate:
		parts := strings.Split(conversationID, "_")
		if len(parts) != 2 {
			return false, errors.New("无效的会话ID")
		}
		friendID := parts[0]
		if friendID == userID {
			friendID = parts[1]
		}
		return s.Repo.Conversation.ToggleFriendPin(ctx, userID, friendID)
	case ConvTypeGroup:
		return s.Repo.Conversation.ToggleGroupPin(ctx, conversationID, userID)
	default:
		return false, errors.New("无效的会话类型")
	}
}

// DeleteConversation 删除会话.
func (s *ConversationService) DeleteConversation(ctx context.Context, userID, conversationType, conversationID string) error {
	switch conversationType {
	case ConvTypePrivate:
		parts := strings.Split(conversationID, "_")
		if len(parts) != 2 {
			return errors.New("无效的会话ID")
		}
		friendID := parts[0]
		if friendID == userID {
			friendID = parts[1]
		}
		if err := s.Repo.Conversation.DeleteFriendRelation(ctx, userID, friendID); err != nil {
			slog.Error("删除好友关系失败", "error", err)
			return err
		}
	case ConvTypeGroup:
		member, err := s.Repo.GroupMember.GetMember(ctx, conversationID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("用户不是群成员")
			}
			slog.Error("获取群成员信息失败", "error", err)
			return err
		}
		if member.Role == RoleOwner {
			return errors.New("群主不能退出群组，请先转让群主或删除群组")
		}
		if err := s.Repo.GroupMember.RemoveMember(ctx, conversationID, userID); err != nil {
			slog.Error("退出群组失败", "error", err)
			return err
		}
	default:
		return errors.New("无效的会话类型")
	}

	return nil
}
