// Package handler 提供 gRPC 请求处理层，负责参数解析、调用 service 并组装响应.
package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AddCollaborator 添加协作者.
func (h *Handler) AddCollaborator(ctx context.Context, req *pb.AddCollaboratorRequest) (*pb.AddCollaboratorResponse, error) {
	if err := h.CollabService.AddCollaborator(ctx, req.GetRepoId(), req.GetOperatorId(), req.GetUserId(), req.GetRole().String()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AddCollaboratorResponse{}, nil
}

// UpdateCollaborator 更新协作者角色.
func (h *Handler) UpdateCollaborator(ctx context.Context, req *pb.UpdateCollaboratorRequest) (*pb.UpdateCollaboratorResponse, error) {
	if err := h.CollabService.UpdateCollaborator(ctx, req.GetRepoId(), req.GetOperatorId(), req.GetUserId(), req.GetRole().String()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateCollaboratorResponse{}, nil
}

// RemoveCollaborator 移除协作者.
func (h *Handler) RemoveCollaborator(ctx context.Context, req *pb.RemoveCollaboratorRequest) (*pb.RemoveCollaboratorResponse, error) {
	if err := h.CollabService.RemoveCollaborator(ctx, req.GetRepoId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RemoveCollaboratorResponse{}, nil
}

// ListCollaborators 列出协作者列表.
func (h *Handler) ListCollaborators(ctx context.Context, req *pb.ListCollaboratorsRequest) (*pb.ListCollaboratorsResponse, error) {
	collaborators, err := h.CollabService.ListCollaborators(ctx, req.GetRepoId(), req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCollabs := make([]*pb.Collaborator, 0, len(collaborators))
	for _, c := range collaborators {
		role := pb.CollaboratorRole_value[c.Role]
		pbCollabs = append(pbCollabs, &pb.Collaborator{
			RepoId: c.RepoID,
			UserId: c.UserID,
			Role:   pb.CollaboratorRole(role),
		})
	}

	return &pb.ListCollaboratorsResponse{
		Collaborators: pbCollabs,
	}, nil
}
