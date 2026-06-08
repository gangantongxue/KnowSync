package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CollaboratorService 协作者业务服务.
type CollaboratorService struct {
	collabRepo CollaboratorRepository
	repoRepo   RepoRepository
}

// NewCollaboratorService 创建协作者服务.
func NewCollaboratorService(
	collabRepo CollaboratorRepository,
	repoRepo RepoRepository,
) *CollaboratorService {
	return &CollaboratorService{
		collabRepo: collabRepo,
		repoRepo:   repoRepo,
	}
}

// AddCollaborator 添加协作者，仅 ADMIN 可操作.
func (s *CollaboratorService) AddCollaborator(ctx context.Context, repoID, operatorID, userID, role string) error {
	if err := CheckRepoPermission(ctx, s.repoRepo, s.collabRepo, nil, repoID, operatorID, "ADMIN"); err != nil {
		return err
	}

	c := &schema.Collaborator{
		RepoID: repoID,
		UserID: userID,
		Role:   role,
	}
	if err := s.collabRepo.AddCollaborator(ctx, c); err != nil {
		slog.Error("添加协作者失败", "error", err)
		return errors.New("添加协作者失败")
	}

	return nil
}

// UpdateCollaborator 更新协作者角色，仅 ADMIN 可操作.
func (s *CollaboratorService) UpdateCollaborator(ctx context.Context, repoID, operatorID, userID, role string) error {
	if err := CheckRepoPermission(ctx, s.repoRepo, s.collabRepo, nil, repoID, operatorID, "ADMIN"); err != nil {
		return err
	}

	if err := s.collabRepo.UpdateCollaborator(ctx, repoID, userID, role); err != nil {
		slog.Error("更新协作者失败", "error", err)
		return errors.New("更新协作者失败")
	}

	return nil
}

// RemoveCollaborator 移除协作者，仅 ADMIN 可操作.
func (s *CollaboratorService) RemoveCollaborator(ctx context.Context, repoID, operatorID, userID string) error {
	if err := CheckRepoPermission(ctx, s.repoRepo, s.collabRepo, nil, repoID, operatorID, "ADMIN"); err != nil {
		return err
	}

	if err := s.collabRepo.RemoveCollaborator(ctx, repoID, userID); err != nil {
		slog.Error("移除协作者失败", "error", err)
		return errors.New("移除协作者失败")
	}

	return nil
}

// ListCollaborators 列出协作者列表，需要 ADMIN 或 DEVELOPER 权限.
func (s *CollaboratorService) ListCollaborators(ctx context.Context, repoID, userID string) ([]schema.Collaborator, error) {
	if err := CheckRepoPermission(ctx, s.repoRepo, s.collabRepo, nil, repoID, userID, "ADMIN", "DEVELOPER"); err != nil {
		return nil, err
	}

	collaborators, err := s.collabRepo.ListCollaborators(ctx, repoID)
	if err != nil {
		slog.Error("查询协作者列表失败", "error", err)
		return nil, errors.New("查询协作者列表失败")
	}

	return collaborators, nil
}
