package service

import (
	"context"
	"log/slog"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// AddCollaborator 添加协作者，仅 ADMIN 可操作
func (s *Service) AddCollaborator(ctx context.Context, req *pb.AddCollaboratorRequest) (*pb.AddCollaboratorResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.OperatorId, "ADMIN"); err != nil {
		return &pb.AddCollaboratorResponse{Success: false, Msg: err.Error()}, nil
	}

	c := &schema.Collaborator{
		RepoID: req.RepoId,
		UserID: req.UserId,
		Role:   req.Role.String(),
	}
	if err := s.Repository.AddCollaborator(ctx, c); err != nil {
		slog.Error("添加协作者失败", "error", err)
		return &pb.AddCollaboratorResponse{Success: false, Msg: "添加协作者失败"}, nil
	}

	return &pb.AddCollaboratorResponse{Success: true}, nil
}

// UpdateCollaborator 更新协作者角色，仅 ADMIN 可操作
func (s *Service) UpdateCollaborator(ctx context.Context, req *pb.UpdateCollaboratorRequest) (*pb.UpdateCollaboratorResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.OperatorId, "ADMIN"); err != nil {
		return &pb.UpdateCollaboratorResponse{Success: false, Msg: err.Error()}, nil
	}

	if err := s.Repository.UpdateCollaborator(ctx, req.RepoId, req.UserId, req.Role.String()); err != nil {
		slog.Error("更新协作者失败", "error", err)
		return &pb.UpdateCollaboratorResponse{Success: false, Msg: "更新协作者失败"}, nil
	}

	return &pb.UpdateCollaboratorResponse{Success: true}, nil
}

// RemoveCollaborator 移除协作者，仅 ADMIN 可操作
func (s *Service) RemoveCollaborator(ctx context.Context, req *pb.RemoveCollaboratorRequest) (*pb.RemoveCollaboratorResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.OperatorId, "ADMIN"); err != nil {
		return &pb.RemoveCollaboratorResponse{Success: false, Msg: err.Error()}, nil
	}

	if err := s.Repository.RemoveCollaborator(ctx, req.RepoId, req.UserId); err != nil {
		slog.Error("移除协作者失败", "error", err)
		return &pb.RemoveCollaboratorResponse{Success: false, Msg: "移除协作者失败"}, nil
	}

	return &pb.RemoveCollaboratorResponse{Success: true}, nil
}

// ListCollaborators 列出协作者列表，需要 ADMIN 或 DEVELOPER 权限
func (s *Service) ListCollaborators(ctx context.Context, req *pb.ListCollaboratorsRequest) (*pb.ListCollaboratorsResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "ADMIN", "DEVELOPER"); err != nil {
		return &pb.ListCollaboratorsResponse{Success: false, Collaborators: []*pb.Collaborator{}}, nil
	}

	collaborators, err := s.Repository.ListCollaborators(ctx, req.RepoId)
	if err != nil {
		slog.Error("查询协作者列表失败", "error", err)
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
