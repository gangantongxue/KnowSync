package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CreateGroup 创建群组
func (h *Handler) CreateGroup(ctx context.Context, req *pb.CreateGroupReq) (*pb.CreateGroupResp, error) {
	group, err := h.Service.CreateGroup(ctx, req.GetUserId(), req.GetName(), req.GetAvatar(), req.GetMemberIds())
	if err != nil {
		return &pb.CreateGroupResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.CreateGroupResp{
		Success: true,
		Msg:     "创建成功",
		Group: &pb.Group{
			Id:        group.ID,
			Name:      group.Name,
			Avatar:    group.Avatar,
			OwnerId:   group.OwnerID,
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
		},
	}, nil
}

// GetGroupInfo 获取群信息
func (h *Handler) GetGroupInfo(ctx context.Context, req *pb.GetGroupInfoReq) (*pb.GetGroupInfoResp, error) {
	group, isMember, err := h.Service.GetGroupInfo(ctx, req.GetGroupId(), req.GetUserId())
	if err != nil {
		return &pb.GetGroupInfoResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.GetGroupInfoResp{
		Success: true,
		Msg:     "获取成功",
		Group: &pb.Group{
			Id:        group.ID,
			Name:      group.Name,
			Avatar:    group.Avatar,
			OwnerId:   group.OwnerID,
			CreatedAt: group.CreatedAt,
			UpdatedAt: group.UpdatedAt,
		},
		IsMember: isMember,
	}, nil
}

// UpdateGroup 更新群信息
func (h *Handler) UpdateGroup(ctx context.Context, req *pb.UpdateGroupReq) (*pb.UpdateGroupResp, error) {
	if err := h.Service.UpdateGroup(ctx, req.GetGroupId(), req.GetUserId(), req.GetName(), req.GetAvatar()); err != nil {
		return &pb.UpdateGroupResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.UpdateGroupResp{
		Success: true,
		Msg:     "更新成功",
	}, nil
}

// AddMembers 添加成员
func (h *Handler) AddMembers(ctx context.Context, req *pb.AddMembersReq) (*pb.AddMembersResp, error) {
	if err := h.Service.AddMembers(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetMemberIds()); err != nil {
		return &pb.AddMembersResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.AddMembersResp{
		Success: true,
		Msg:     "添加成功",
	}, nil
}

// RemoveMember 移除成员
func (h *Handler) RemoveMember(ctx context.Context, req *pb.RemoveMemberReq) (*pb.RemoveMemberResp, error) {
	if err := h.Service.RemoveMember(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return &pb.RemoveMemberResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.RemoveMemberResp{
		Success: true,
		Msg:     "移除成功",
	}, nil
}

// LeaveGroup 退出群组
func (h *Handler) LeaveGroup(ctx context.Context, req *pb.LeaveGroupReq) (*pb.LeaveGroupResp, error) {
	if err := h.Service.LeaveGroup(ctx, req.GetGroupId(), req.GetUserId()); err != nil {
		return &pb.LeaveGroupResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.LeaveGroupResp{
		Success: true,
		Msg:     "退出成功",
	}, nil
}

// TransferOwnership 转让群主
func (h *Handler) TransferOwnership(ctx context.Context, req *pb.TransferOwnershipReq) (*pb.TransferOwnershipResp, error) {
	if err := h.Service.TransferOwnership(ctx, req.GetGroupId(), req.GetCurrentOwnerId(), req.GetNewOwnerId()); err != nil {
		return &pb.TransferOwnershipResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.TransferOwnershipResp{
		Success: true,
		Msg:     "转让成功",
	}, nil
}

// SetAdmin 设置管理员
func (h *Handler) SetAdmin(ctx context.Context, req *pb.SetAdminReq) (*pb.SetAdminResp, error) {
	if err := h.Service.SetAdmin(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return &pb.SetAdminResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.SetAdminResp{
		Success: true,
		Msg:     "设置成功",
	}, nil
}

// RemoveAdmin 移除管理员
func (h *Handler) RemoveAdmin(ctx context.Context, req *pb.RemoveAdminReq) (*pb.RemoveAdminResp, error) {
	if err := h.Service.RemoveAdmin(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return &pb.RemoveAdminResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.RemoveAdminResp{
		Success: true,
		Msg:     "移除成功",
	}, nil
}

// GetGroupMembers 获取群成员
func (h *Handler) GetGroupMembers(ctx context.Context, req *pb.GetGroupMembersReq) (*pb.GetGroupMembersResp, error) {
	members, err := h.Service.GetGroupMembers(ctx, req.GetGroupId(), req.GetUserId())
	if err != nil {
		return &pb.GetGroupMembersResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	pbMembers := make([]*pb.GroupMember, len(members))
	for i, m := range members {
		pbMembers[i] = &pb.GroupMember{
			Id:       m.ID,
			GroupId:  m.GroupID,
			UserId:   m.UserID,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		}
	}
	return &pb.GetGroupMembersResp{
		Success: true,
		Msg:     "获取成功",
		Members: pbMembers,
	}, nil
}

// GetUserGroups 获取用户群列表
func (h *Handler) GetUserGroups(ctx context.Context, req *pb.GetUserGroupsReq) (*pb.GetUserGroupsResp, error) {
	groups, err := h.Service.GetUserGroups(ctx, req.GetUserId())
	if err != nil {
		return &pb.GetUserGroupsResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	pbGroups := make([]*pb.Group, len(groups))
	for i, g := range groups {
		pbGroups[i] = &pb.Group{
			Id:        g.ID,
			Name:      g.Name,
			Avatar:    g.Avatar,
			OwnerId:   g.OwnerID,
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
		}
	}
	return &pb.GetUserGroupsResp{
		Success: true,
		Msg:     "获取成功",
		Groups:  pbGroups,
	}, nil
}
