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

// CreateGroup 创建群组
func (r *GroupRepository) CreateGroup(ctx context.Context, group *schema.Group) error {
	return r.DB.WithContext(ctx).Create(group).Error
}

// GetGroupByID 根据ID获取群组
func (r *GroupRepository) GetGroupByID(ctx context.Context, id string) (*schema.Group, error) {
	var group schema.Group
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup 更新群组信息
func (r *GroupRepository) UpdateGroup(ctx context.Context, group *schema.Group) error {
	return r.DB.WithContext(ctx).Save(group).Error
}

// DeleteGroup 删除群组
func (r *GroupRepository) DeleteGroup(ctx context.Context, groupID string) error {
	return r.DB.WithContext(ctx).Where("id = ?", groupID).Delete(&schema.Group{}).Error
}

// UserGroupInfo 用户群组信息（包含成员角色）
type UserGroupInfo struct {
	schema.Group
	Role string
}

// GetUserGroups 获取用户加入的所有群组（包含成员角色）
func (r *GroupRepository) GetUserGroups(ctx context.Context, userID string) ([]UserGroupInfo, error) {
	var results []UserGroupInfo
	if err := r.DB.WithContext(ctx).
		Table("groups").
		Select("groups.*, group_members.role").
		Joins("JOIN group_members ON group_members.group_id = groups.id").
		Where("group_members.user_id = ?", userID).
		Order("groups.created_at DESC").
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GroupMemberRepository 群组成员仓库
type GroupMemberRepository struct {
	DB *gorm.DB
}

// NewGroupMemberRepository 创建群组成员仓库
func NewGroupMemberRepository(db *gorm.DB) *GroupMemberRepository {
	return &GroupMemberRepository{DB: db}
}

// AddMember 添加成员
func (r *GroupMemberRepository) AddMember(ctx context.Context, member *schema.GroupMember) error {
	return r.DB.WithContext(ctx).Create(member).Error
}

// RemoveMember 移除成员
func (r *GroupMemberRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	return r.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&schema.GroupMember{}).Error
}

// UpdateMemberRole 更新成员角色
func (r *GroupMemberRepository) UpdateMemberRole(ctx context.Context, groupID, userID, role string) error {
	return r.DB.WithContext(ctx).
		Model(&schema.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role", role).Error
}

// GetMembers 获取所有成员
func (r *GroupMemberRepository) GetMembers(ctx context.Context, groupID string) ([]schema.GroupMember, error) {
	var members []schema.GroupMember
	if err := r.DB.WithContext(ctx).
		Where("group_id = ?", groupID).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// GetMember 获取单个成员
func (r *GroupMemberRepository) GetMember(ctx context.Context, groupID, userID string) (*schema.GroupMember, error) {
	var member schema.GroupMember
	if err := r.DB.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// CountMembers 统计成员数
func (r *GroupMemberRepository) CountMembers(ctx context.Context, groupID string) (int64, error) {
	var count int64
	if err := r.DB.WithContext(ctx).
		Model(&schema.GroupMember{}).
		Where("group_id = ?", groupID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
