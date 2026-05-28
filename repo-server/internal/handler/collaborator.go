package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// AddCollaborator 添加协作者
func (h *Handler) AddCollaborator(ctx context.Context, req *pb.AddCollaboratorRequest) (*pb.AddCollaboratorResponse, error) {
	if err := h.Service.AddCollaborator(ctx, req.GetRepoId(), req.GetOperatorId(), req.GetUserId(), req.GetRole().String()); err != nil {
		return &pb.AddCollaboratorResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.AddCollaboratorResponse{Success: true}, nil
}

// UpdateCollaborator 更新协作者角色
func (h *Handler) UpdateCollaborator(ctx context.Context, req *pb.UpdateCollaboratorRequest) (*pb.UpdateCollaboratorResponse, error) {
	if err := h.Service.UpdateCollaborator(ctx, req.GetRepoId(), req.GetOperatorId(), req.GetUserId(), req.GetRole().String()); err != nil {
		return &pb.UpdateCollaboratorResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.UpdateCollaboratorResponse{Success: true}, nil
}

// RemoveCollaborator 移除协作者
func (h *Handler) RemoveCollaborator(ctx context.Context, req *pb.RemoveCollaboratorRequest) (*pb.RemoveCollaboratorResponse, error) {
	if err := h.Service.RemoveCollaborator(ctx, req.GetRepoId(), req.GetOperatorId(), req.GetUserId()); err != nil {
		return &pb.RemoveCollaboratorResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.RemoveCollaboratorResponse{Success: true}, nil
}

// ListCollaborators 列出协作者列表
func (h *Handler) ListCollaborators(ctx context.Context, req *pb.ListCollaboratorsRequest) (*pb.ListCollaboratorsResponse, error) {
	collaborators, err := h.Service.ListCollaborators(ctx, req.GetRepoId(), req.GetUserId())
	if err != nil {
		return &pb.ListCollaboratorsResponse{Success: false, Collaborators: []*pb.Collaborator{}}, nil
	}

	pbCollabs := make([]*pb.Collaborator, 0, len(collaborators))
	for _, c := range collaborators {
		role, _ := pb.CollaboratorRole_value[c.Role]
		pbCollabs = append(pbCollabs, &pb.Collaborator{
			RepoId: c.RepoID,
			UserId: c.UserID,
			Role:   pb.CollaboratorRole(role),
		})
	}

	return &pb.ListCollaboratorsResponse{
		Success:       true,
		Collaborators: pbCollabs,
	}, nil
}
