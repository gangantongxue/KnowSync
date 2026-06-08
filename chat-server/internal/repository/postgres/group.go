package postgres

import (
	"context"

	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// GroupRepo 群组仓库实现.
type GroupRepo struct {
	db *gorm.DB
}

// NewGroupRepo 创建群组仓库.
func NewGroupRepo(db *gorm.DB) *GroupRepo {
	return &GroupRepo{db: db}
}

// CreateGroup 创建群组.
func (r *GroupRepo) CreateGroup(ctx context.Context, group *schema.Group) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroup 获取群组.
func (r *GroupRepo) GetGroup(ctx context.Context, id string) (*schema.Group, error) {
	var group schema.Group
	if err := r.db.WithContext(ctx).First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup 更新群组.
func (r *GroupRepo) UpdateGroup(ctx context.Context, group *schema.Group) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteGroup 删除群组.
func (r *GroupRepo) DeleteGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&schema.Group{}, "id = ?", id).Error
}

// AddGroupMember 添加群成员.
func (r *GroupRepo) AddGroupMember(ctx context.Context, member *schema.GroupMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveGroupMember 移除群成员.
func (r *GroupRepo) RemoveGroupMember(ctx context.Context, groupID, userID string) error {
	return r.db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&schema.GroupMember{}).Error
}

// ListGroupMembers 列出群成员.
func (r *GroupRepo) ListGroupMembers(ctx context.Context, groupID string) ([]*schema.GroupMember, error) {
	var members []*schema.GroupMember
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}
