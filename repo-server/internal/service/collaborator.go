package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// AddCollaborator 添加协作者，仅 ADMIN 可操作
func (s *Service) AddCollaborator(ctx context.Context, repoID string, operatorID string, userID string, role string) error {
	if err := s.CheckRepoPermission(ctx, repoID, operatorID, "ADMIN"); err != nil {
		return err
	}

	c := &schema.Collaborator{
		RepoID: repoID,
		UserID: userID,
		Role:   role,
	}
	if err := s.Repository.AddCollaborator(ctx, c); err != nil {
		slog.Error("添加协作者失败", "error", err)
		return fmt.Errorf("添加协作者失败")
	}

	return nil
}

// UpdateCollaborator 更新协作者角色，仅 ADMIN 可操作
func (s *Service) UpdateCollaborator(ctx context.Context, repoID string, operatorID string, userID string, role string) error {
	if err := s.CheckRepoPermission(ctx, repoID, operatorID, "ADMIN"); err != nil {
		return err
	}

	if err := s.Repository.UpdateCollaborator(ctx, repoID, userID, role); err != nil {
		slog.Error("更新协作者失败", "error", err)
		return fmt.Errorf("更新协作者失败")
	}

	return nil
}

// RemoveCollaborator 移除协作者，仅 ADMIN 可操作
func (s *Service) RemoveCollaborator(ctx context.Context, repoID string, operatorID string, userID string) error {
	if err := s.CheckRepoPermission(ctx, repoID, operatorID, "ADMIN"); err != nil {
		return err
	}

	if err := s.Repository.RemoveCollaborator(ctx, repoID, userID); err != nil {
		slog.Error("移除协作者失败", "error", err)
		return fmt.Errorf("移除协作者失败")
	}

	return nil
}

// ListCollaborators 列出协作者列表，需要 ADMIN 或 DEVELOPER 权限
func (s *Service) ListCollaborators(ctx context.Context, repoID string, userID string) ([]schema.Collaborator, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "ADMIN", "DEVELOPER"); err != nil {
		return nil, err
	}

	collaborators, err := s.Repository.ListCollaborators(ctx, repoID)
	if err != nil {
		slog.Error("查询协作者列表失败", "error", err)
		return nil, fmt.Errorf("查询协作者列表失败")
	}

	return collaborators, nil
}
