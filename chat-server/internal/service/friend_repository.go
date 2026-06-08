package service

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// FriendRepository 好友数据访问接口（消费者定义）.
type FriendRepository interface {
	AddFriend(ctx context.Context, friend *schema.Friend) error
	RemoveFriend(ctx context.Context, userID, friendID string) error
	IsFriend(ctx context.Context, userID, friendID string) (bool, error)
	ListFriends(ctx context.Context, userID string) ([]*schema.Friend, error)
}
