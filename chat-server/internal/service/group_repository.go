package service

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
)

// GroupRepository 群组数据访问接口（消费者定义）.
type GroupRepository interface {
	CreateGroup(ctx context.Context, group *schema.Group) error
	GetGroup(ctx context.Context, id string) (*schema.Group, error)
	UpdateGroup(ctx context.Context, group *schema.Group) error
	DeleteGroup(ctx context.Context, id string) error
	AddGroupMember(ctx context.Context, member *schema.GroupMember) error
	RemoveGroupMember(ctx context.Context, groupID, userID string) error
	ListGroupMembers(ctx context.Context, groupID string) ([]*schema.GroupMember, error)
}
