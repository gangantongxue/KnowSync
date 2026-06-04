package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// FollowRepo 关注知识库，仅支持公开知识库，且不能关注自己的知识库
func (s *Service) FollowRepo(ctx context.Context, repoID, userID string) error {
	repo, err := s.Repository.GetRepo(ctx, repoID)
	if err != nil {
		slog.Error("查询知识库失败", "error", err)
		return fmt.Errorf("知识库不存在")
	}

	if repo.Visibility != "PUBLIC" {
		return fmt.Errorf("仅支持关注公开知识库")
	}

	if repo.OwnerID == userID {
		return fmt.Errorf("不能关注自己的知识库")
	}

	if err := s.Repository.FollowRepo(ctx, userID, repoID); err != nil {
		slog.Error("关注知识库失败", "error", err)
		return fmt.Errorf("关注知识库失败")
	}

	return nil
}

// UnfollowRepo 取消关注知识库
func (s *Service) UnfollowRepo(ctx context.Context, repoID, userID string) error {
	if err := s.Repository.UnfollowRepo(ctx, userID, repoID); err != nil {
		slog.Error("取消关注失败", "error", err)
		return fmt.Errorf("取消关注失败")
	}

	return nil
}

// ListFollowedRepos 获取用户关注的知识库列表
func (s *Service) ListFollowedRepos(ctx context.Context, userID string) ([]schema.Repo, error) {
	repoIDs, err := s.Repository.ListFollowedRepoIDs(ctx, userID)
	if err != nil {
		slog.Error("查询关注列表失败", "error", err)
		return nil, fmt.Errorf("查询关注列表失败")
	}

	if len(repoIDs) == 0 {
		return nil, nil
	}

	repos, err := s.Repository.ListReposByIDs(ctx, repoIDs)
	if err != nil {
		slog.Error("查询关注知识库详情失败", "error", err)
		return nil, fmt.Errorf("查询关注列表失败")
	}

	return repos, nil
}
