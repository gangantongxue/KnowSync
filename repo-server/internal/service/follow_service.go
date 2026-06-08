package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// FollowService 关注业务服务.
type FollowService struct {
	followRepo FollowRepository
	repoRepo   RepoRepository
}

// NewFollowService 创建关注服务.
func NewFollowService(
	followRepo FollowRepository,
	repoRepo RepoRepository,
) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		repoRepo:   repoRepo,
	}
}

// FollowRepo 关注知识库，仅支持公开知识库，且不能关注自己的知识库.
func (s *FollowService) FollowRepo(ctx context.Context, repoID, userID string) error {
	repo, err := s.repoRepo.GetRepo(ctx, repoID)
	if err != nil {
		slog.Error("查询知识库失败", "error", err)
		return errors.New("知识库不存在")
	}

	if repo.Visibility != "PUBLIC" {
		return errors.New("仅支持关注公开知识库")
	}

	if repo.OwnerID == userID {
		return errors.New("不能关注自己的知识库")
	}

	if err := s.followRepo.FollowRepo(ctx, userID, repoID); err != nil {
		slog.Error("关注知识库失败", "error", err)
		return errors.New("关注知识库失败")
	}

	return nil
}

// UnfollowRepo 取消关注知识库.
func (s *FollowService) UnfollowRepo(ctx context.Context, repoID, userID string) error {
	if err := s.followRepo.UnfollowRepo(ctx, userID, repoID); err != nil {
		slog.Error("取消关注失败", "error", err)
		return errors.New("取消关注失败")
	}

	return nil
}

// ListFollowedRepos 获取用户关注的知识库列表.
func (s *FollowService) ListFollowedRepos(ctx context.Context, userID string) ([]schema.Repo, error) {
	repoIDs, err := s.followRepo.ListFollowedRepoIDs(ctx, userID)
	if err != nil {
		slog.Error("查询关注列表失败", "error", err)
		return nil, errors.New("查询关注列表失败")
	}

	if len(repoIDs) == 0 {
		return nil, nil
	}

	repos, err := s.repoRepo.ListReposByIDs(ctx, repoIDs)
	if err != nil {
		slog.Error("查询关注知识库详情失败", "error", err)
		return nil, errors.New("查询关注列表失败")
	}

	return repos, nil
}

// CountFollowers 获取知识库的关注总数.
func (s *FollowService) CountFollowers(ctx context.Context, repoID string) (int64, error) {
	return s.followRepo.CountFollowers(ctx, repoID)
}

// BatchCountFollowers 批量获取多个知识库的关注总数.
func (s *FollowService) BatchCountFollowers(ctx context.Context, repoIDs []string) (map[string]int64, error) {
	return s.followRepo.BatchCountFollowers(ctx, repoIDs)
}

// IsFollowing 检查用户是否已关注知识库.
func (s *FollowService) IsFollowing(ctx context.Context, repoID, userID string) (bool, error) {
	return s.followRepo.IsFollowing(ctx, userID, repoID)
}
