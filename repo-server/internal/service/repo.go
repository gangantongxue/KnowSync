package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CreateRepo 创建知识库，同时将创建者添加为 ADMIN 角色协作者.
func (s *Service) CreateRepo(ctx context.Context, ownerID, name, visibility, desc string) (*schema.Repo, error) {
	repo := &schema.Repo{
		OwnerID:     ownerID,
		Name:        name,
		Visibility:  visibility,
		Description: desc,
	}
	if err := s.Repository.CreateRepo(ctx, repo); err != nil {
		slog.Error("创建知识库失败", "error", err)
		return nil, errors.New("创建知识库失败")
	}

	collaborator := &schema.Collaborator{
		RepoID: repo.ID,
		UserID: ownerID,
		Role:   "ADMIN",
	}
	if err := s.Repository.AddCollaborator(ctx, collaborator); err != nil {
		slog.Error("添加所有者协作者失败", "error", err)
	}

	return repo, nil
}

// GetRepo 获取知识库详情.
func (s *Service) GetRepo(ctx context.Context, repoID, userID string) (*schema.Repo, string, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "VIEWER"); err != nil {
		return nil, "", err
	}

	repo, err := s.Repository.GetRepo(ctx, repoID)
	if err != nil {
		return nil, "", errors.New("知识库不存在")
	}

	myRole := ""
	roleStr, err := s.Repository.GetUserRole(ctx, repoID, userID)
	if err == nil {
		myRole = roleStr
	}

	return repo, myRole, nil
}

// UpdateRepo 更新知识库信息，仅 ADMIN 可操作.
func (s *Service) UpdateRepo(ctx context.Context, repoID, userID, name, desc, visibility string) (*schema.Repo, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "ADMIN"); err != nil {
		return nil, err
	}

	repo, err := s.Repository.GetRepo(ctx, repoID)
	if err != nil {
		return nil, errors.New("知识库不存在")
	}

	if name != "" {
		repo.Name = name
	}
	if desc != "" {
		repo.Description = desc
	}
	if visibility != "" {
		repo.Visibility = visibility
	}

	if err := s.Repository.UpdateRepo(ctx, repo); err != nil {
		slog.Error("更新知识库失败", "error", err)
		return nil, errors.New("更新知识库失败")
	}

	return repo, nil
}

// DeleteRepo 软删除知识库，仅 ADMIN 可操作.
func (s *Service) DeleteRepo(ctx context.Context, repoID, userID string) error {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "ADMIN"); err != nil {
		return err
	}

	if err := s.Repository.DeleteRepo(ctx, repoID); err != nil {
		slog.Error("删除知识库失败", "error", err)
		return errors.New("删除知识库失败")
	}

	return nil
}

// ListPublicRepos 获取所有公开知识库列表.
func (s *Service) ListPublicRepos(ctx context.Context) ([]schema.Repo, error) {
	repos, err := s.Repository.ListPublicRepos(ctx)
	if err != nil {
		slog.Error("查询公开知识库列表失败", "error", err)
		return nil, errors.New("查询公开知识库列表失败")
	}
	return repos, nil
}

// ListUserRepos 获取用户参与的所有知识库列表.
func (s *Service) ListUserRepos(ctx context.Context, userID string) ([]schema.Repo, error) {
	repoIDs, err := s.Repository.ListUserRepoIDs(ctx, userID)
	if err != nil {
		slog.Error("查询用户知识库列表失败", "error", err)
		return nil, errors.New("查询知识库列表失败")
	}

	repos, err := s.Repository.ListReposByIDs(ctx, repoIDs)
	if err != nil {
		return nil, errors.New("查询知识库列表失败")
	}

	return repos, nil
}
