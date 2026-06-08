// Package handler 提供 gRPC 消息处理逻辑.
package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateGroup 创建群组.
func (h *Handler) CreateGroup(ctx context.Context, req *pb.CreateGroupReq) (*pb.CreateGroupResp, error) {
	group, err := h.GroupService.CreateGroup(ctx, req.GetUserId(), req.GetName(), req.GetAvatar(), req.GetMemberIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateGroupResp{
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

// GetGroupInfo 获取群信息.
func (h *Handler) GetGroupInfo(ctx context.Context, req *pb.GetGroupInfoReq) (*pb.GetGroupInfoResp, error) {
	group, isMember, err := h.GroupService.GetGroupInfo(ctx, req.GetGroupId(), req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetGroupInfoResp{
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

// UpdateGroup 更新群信息.
func (h *Handler) UpdateGroup(ctx context.Context, req *pb.UpdateGroupReq) (*pb.UpdateGroupResp, error) {
	if err := h.GroupService.UpdateGroup(ctx, req.GetGroupId(), req.GetUserId(), req.GetName(), req.GetAvatar()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateGroupResp{}, nil
}

// AddMembers 添加成员.
func (h *Handler) AddMembers(ctx context.Context, req *pb.AddMembersReq) (*pb.AddMembersResp, error) {
	if err := h.GroupService.AddMembers(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetMemberIds()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AddMembersResp{}, nil
}

// RemoveMember 移除成员.
func (h *Handler) RemoveMember(ctx context.Context, req *pb.RemoveMemberReq) (*pb.RemoveMemberResp, error) {
	if err := h.GroupService.RemoveMember(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RemoveMemberResp{}, nil
}

// LeaveGroup 退出群组.
func (h *Handler) LeaveGroup(ctx context.Context, req *pb.LeaveGroupReq) (*pb.LeaveGroupResp, error) {
	if err := h.GroupService.LeaveGroup(ctx, req.GetGroupId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LeaveGroupResp{}, nil
}

// TransferOwnership 转让群主.
func (h *Handler) TransferOwnership(ctx context.Context, req *pb.TransferOwnershipReq) (*pb.TransferOwnershipResp, error) {
	if err := h.GroupService.TransferOwnership(ctx, req.GetGroupId(), req.GetCurrentOwnerId(), req.GetNewOwnerId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TransferOwnershipResp{}, nil
}

// SetAdmin 设置管理员.
func (h *Handler) SetAdmin(ctx context.Context, req *pb.SetAdminReq) (*pb.SetAdminResp, error) {
	if err := h.GroupService.SetAdmin(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.SetAdminResp{}, nil
}

// RemoveAdmin 移除管理员.
func (h *Handler) RemoveAdmin(ctx context.Context, req *pb.RemoveAdminReq) (*pb.RemoveAdminResp, error) {
	if err := h.GroupService.RemoveAdmin(ctx, req.GetGroupId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RemoveAdminResp{}, nil
}

// GetGroupMembers 获取群成员.
func (h *Handler) GetGroupMembers(ctx context.Context, req *pb.GetGroupMembersReq) (*pb.GetGroupMembersResp, error) {
	members, err := h.GroupService.GetGroupMembers(ctx, req.GetGroupId(), req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		Members: pbMembers,
	}, nil
}

// GetUserGroups 获取用户群列表.
func (h *Handler) GetUserGroups(ctx context.Context, req *pb.GetUserGroupsReq) (*pb.GetUserGroupsResp, error) {
	groups, err := h.GroupService.GetUserGroups(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
		Groups: pbGroups,
	}, nil
}
