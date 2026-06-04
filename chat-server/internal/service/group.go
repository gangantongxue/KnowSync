package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gangantongxue/knowsync/chat-server/internal/repository"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// CreateGroup 创建群组，添加群主和初始成员.
func (s *Service) CreateGroup(ctx context.Context, ownerID, name, avatar string, memberIDs []string) (*schema.Group, error) {
	now := time.Now().Unix()
	group := &schema.Group{
		Name:      name,
		Avatar:    avatar,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.Repo.Group.CreateGroup(ctx, group); err != nil {
		slog.Error("创建群组失败", "error", err)
		return nil, err
	}

	ownerMember := &schema.GroupMember{
		GroupID:  group.ID,
		UserID:   ownerID,
		Role:     RoleOwner,
		JoinedAt: now,
	}
	if err := s.Repo.GroupMember.AddMember(ctx, ownerMember); err != nil {
		slog.Error("添加群主失败", "error", err)
		return nil, err
	}

	for _, memberID := range memberIDs {
		if memberID == ownerID {
			continue
		}
		member := &schema.GroupMember{
			GroupID:  group.ID,
			UserID:   memberID,
			Role:     "member",
			JoinedAt: now,
		}
		if err := s.Repo.GroupMember.AddMember(ctx, member); err != nil {
			slog.Error("添加成员失败", "error", err, "user_id", memberID)
			return nil, err
		}
	}

	return group, nil
}

// GetGroupInfo 获取群组信息及用户是否在群中.
func (s *Service) GetGroupInfo(ctx context.Context, groupID, userID string) (*schema.Group, bool, error) {
	group, err := s.Repo.Group.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, errors.New("群组不存在")
		}
		slog.Error("获取群组失败", "error", err)
		return nil, false, err
	}

	_, err = s.Repo.GroupMember.GetMember(ctx, groupID, userID)
	isMember := true
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isMember = false
		} else {
			slog.Error("检查群成员失败", "error", err)
			return nil, false, err
		}
	}

	return group, isMember, nil
}

// UpdateGroup 更新群组信息（仅群主可操作）.
func (s *Service) UpdateGroup(ctx context.Context, groupID, userID, name, avatar string) error {
	group, err := s.Repo.Group.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("群组不存在")
		}
		slog.Error("获取群组失败", "error", err)
		return err
	}

	if group.OwnerID != userID {
		return errors.New("仅群主可以更新群组信息")
	}

	group.Name = name
	group.Avatar = avatar
	group.UpdatedAt = time.Now().Unix()

	if err := s.Repo.Group.UpdateGroup(ctx, group); err != nil {
		slog.Error("更新群组失败", "error", err)
		return err
	}

	return nil
}

// AddMembers 添加成员（群主/管理员可操作）.
func (s *Service) AddMembers(ctx context.Context, groupID, operatorID string, memberIDs []string) error {
	operator, err := s.Repo.GroupMember.GetMember(ctx, groupID, operatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("操作者不是群成员")
		}
		slog.Error("获取操作者信息失败", "error", err)
		return err
	}

	if operator.Role != RoleOwner && operator.Role != RoleAdmin {
		return errors.New("仅群主和管理员可以添加成员")
	}

	now := time.Now().Unix()
	for _, memberID := range memberIDs {
		_, err := s.Repo.GroupMember.GetMember(ctx, groupID, memberID)
		if err == nil {
			slog.Warn("用户已是群成员", "user_id", memberID)
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("检查成员失败", "error", err, "user_id", memberID)
			return err
		}

		member := &schema.GroupMember{
			GroupID:  groupID,
			UserID:   memberID,
			Role:     "member",
			JoinedAt: now,
		}
		if err := s.Repo.GroupMember.AddMember(ctx, member); err != nil {
			slog.Error("添加成员失败", "error", err, "user_id", memberID)
			return err
		}
	}

	return nil
}

// RemoveMember 移除成员.
func (s *Service) RemoveMember(ctx context.Context, groupID, operatorID, targetUserID string) error {
	if operatorID == targetUserID {
		return errors.New("不能移除自己，请使用退出群组功能")
	}

	operator, err := s.Repo.GroupMember.GetMember(ctx, groupID, operatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("操作者不是群成员")
		}
		slog.Error("获取操作者信息失败", "error", err)
		return err
	}

	target, err := s.Repo.GroupMember.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("目标用户不是群成员")
		}
		slog.Error("获取目标用户信息失败", "error", err)
		return err
	}

	if operator.Role != RoleOwner && operator.Role != RoleAdmin {
		return errors.New("仅群主和管理员可以移除成员")
	}

	if operator.Role == RoleAdmin && (target.Role == RoleOwner || target.Role == RoleAdmin) {
		return errors.New("管理员不能移除群主或其他管理员")
	}

	if err := s.Repo.GroupMember.RemoveMember(ctx, groupID, targetUserID); err != nil {
		slog.Error("移除成员失败", "error", err)
		return err
	}

	return nil
}

// LeaveGroup 退出群组.
func (s *Service) LeaveGroup(ctx context.Context, groupID, userID string) error {
	member, err := s.Repo.GroupMember.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不是群成员")
		}
		slog.Error("获取成员信息失败", "error", err)
		return err
	}

	if member.Role == RoleOwner {
		return errors.New("群主不能退出群组，请先转让群主或删除群组")
	}

	if err := s.Repo.GroupMember.RemoveMember(ctx, groupID, userID); err != nil {
		slog.Error("退出群组失败", "error", err)
		return err
	}

	return nil
}

// TransferOwnership 转让群主.
func (s *Service) TransferOwnership(ctx context.Context, groupID, currentOwnerID, newOwnerID string) error {
	if currentOwnerID == newOwnerID {
		return errors.New("不能转让给自己")
	}

	currentOwner, err := s.Repo.GroupMember.GetMember(ctx, groupID, currentOwnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("当前用户不是群成员")
		}
		slog.Error("获取当前群主信息失败", "error", err)
		return err
	}
	if currentOwner.Role != RoleOwner {
		return errors.New("仅群主可以转让群组")
	}

	_, err = s.Repo.GroupMember.GetMember(ctx, groupID, newOwnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("新群主不是群成员")
		}
		slog.Error("获取新群主信息失败", "error", err)
		return err
	}

	group, err := s.Repo.Group.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Error("获取群组失败", "error", err)
		return err
	}
	group.OwnerID = newOwnerID
	group.UpdatedAt = time.Now().Unix()
	if err := s.Repo.Group.UpdateGroup(ctx, group); err != nil {
		slog.Error("更新群组失败", "error", err)
		return err
	}

	if err := s.Repo.GroupMember.UpdateMemberRole(ctx, groupID, currentOwnerID, "member"); err != nil {
		slog.Error("更新旧群主角色失败", "error", err)
		return err
	}

	if err := s.Repo.GroupMember.UpdateMemberRole(ctx, groupID, newOwnerID, RoleOwner); err != nil {
		slog.Error("更新新群主角色失败", "error", err)
		return err
	}

	return nil
}

// SetAdmin 设置管理员（仅群主可操作）.
func (s *Service) SetAdmin(ctx context.Context, groupID, operatorID, targetUserID string) error {
	if operatorID == targetUserID {
		return errors.New("不能设置自己为管理员")
	}

	operator, err := s.Repo.GroupMember.GetMember(ctx, groupID, operatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("操作者不是群成员")
		}
		slog.Error("获取操作者信息失败", "error", err)
		return err
	}
	if operator.Role != RoleOwner {
		return errors.New("仅群主可以设置管理员")
	}

	target, err := s.Repo.GroupMember.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("目标用户不是群成员")
		}
		slog.Error("获取目标用户信息失败", "error", err)
		return err
	}

	if target.Role == RoleOwner {
		return errors.New("不能设置群主为管理员")
	}

	if target.Role == RoleAdmin {
		return errors.New("该用户已经是管理员")
	}

	if err := s.Repo.GroupMember.UpdateMemberRole(ctx, groupID, targetUserID, RoleAdmin); err != nil {
		slog.Error("设置管理员失败", "error", err)
		return err
	}

	return nil
}

// RemoveAdmin 移除管理员（仅群主可操作）.
func (s *Service) RemoveAdmin(ctx context.Context, groupID, operatorID, targetUserID string) error {
	if operatorID == targetUserID {
		return errors.New("不能操作自己")
	}

	operator, err := s.Repo.GroupMember.GetMember(ctx, groupID, operatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("操作者不是群成员")
		}
		slog.Error("获取操作者信息失败", "error", err)
		return err
	}
	if operator.Role != RoleOwner {
		return errors.New("仅群主可以移除管理员")
	}

	target, err := s.Repo.GroupMember.GetMember(ctx, groupID, targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("目标用户不是群成员")
		}
		slog.Error("获取目标用户信息失败", "error", err)
		return err
	}

	if target.Role != RoleAdmin {
		return errors.New("目标用户不是管理员")
	}

	if err := s.Repo.GroupMember.UpdateMemberRole(ctx, groupID, targetUserID, "member"); err != nil {
		slog.Error("移除管理员失败", "error", err)
		return err
	}

	return nil
}

// GetGroupMembers 获取群成员列表（必须为群成员）.
func (s *Service) GetGroupMembers(ctx context.Context, groupID, userID string) ([]schema.GroupMember, error) {
	_, err := s.Repo.GroupMember.GetMember(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不是群成员")
		}
		slog.Error("检查群成员失败", "error", err)
		return nil, err
	}

	members, err := s.Repo.GroupMember.GetMembers(ctx, groupID)
	if err != nil {
		slog.Error("获取群成员失败", "error", err)
		return nil, err
	}

	return members, nil
}

// GetUserGroups 获取用户加入的所有群组.
func (s *Service) GetUserGroups(ctx context.Context, userID string) ([]repository.UserGroupInfo, error) {
	groups, err := s.Repo.Group.GetUserGroups(ctx, userID)
	if err != nil {
		slog.Error("获取用户群组失败", "error", err)
		return nil, err
	}
	return groups, nil
}

// DeleteGroup 删除群组（仅群主可操作）.
func (s *Service) DeleteGroup(ctx context.Context, groupID, userID string) error {
	group, err := s.Repo.Group.GetGroupByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("群组不存在")
		}
		slog.Error("获取群组失败", "error", err)
		return err
	}

	if group.OwnerID != userID {
		return errors.New("仅群主可以删除群组")
	}

	members, err := s.Repo.GroupMember.GetMembers(ctx, groupID)
	if err != nil {
		slog.Error("获取群成员列表失败", "error", err)
		return err
	}
	for _, member := range members {
		if err := s.Repo.GroupMember.RemoveMember(ctx, groupID, member.UserID); err != nil {
			slog.Error("删除成员失败", "error", err, "user_id", member.UserID)
			return err
		}
	}

	if err := s.Repo.Group.DeleteGroup(ctx, groupID); err != nil {
		slog.Error("删除群组失败", "error", err)
		return err
	}

	return nil
}
