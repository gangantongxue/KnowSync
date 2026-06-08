package postgres

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// FollowRepo 关注数据访问实现.
type FollowRepo struct {
	db *database.Database
}

// NewFollowRepo 创建关注数据访问实例.
func NewFollowRepo(database *database.Database) *FollowRepo {
	return &FollowRepo{db: database}
}

// FollowRepo 关注知识库.
func (r *FollowRepo) FollowRepo(ctx context.Context, userID, repoID string) error {
	f := &schema.Follow{
		UserID: userID,
		RepoID: repoID,
	}
	return r.db.DB.WithContext(ctx).Where("user_id = ? AND repo_id = ?", userID, repoID).
		FirstOrCreate(f).Error
}

// UnfollowRepo 取消关注知识库.
func (r *FollowRepo) UnfollowRepo(ctx context.Context, userID, repoID string) error {
	return r.db.DB.WithContext(ctx).
		Where("user_id = ? AND repo_id = ?", userID, repoID).
		Delete(&schema.Follow{}).Error
}

// IsFollowing 检查用户是否已关注知识库.
func (r *FollowRepo) IsFollowing(ctx context.Context, userID, repoID string) (bool, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).
		Model(&schema.Follow{}).
		Where("user_id = ? AND repo_id = ?", userID, repoID).
		Count(&count).Error
	return count > 0, err
}

// CountFollowers 获取知识库的关注总数.
func (r *FollowRepo) CountFollowers(ctx context.Context, repoID string) (int64, error) {
	var count int64
	err := r.db.DB.WithContext(ctx).
		Model(&schema.Follow{}).
		Where("repo_id = ?", repoID).
		Count(&count).Error
	return count, err
}

// BatchCountFollowers 批量获取多个知识库的关注总数.
func (r *FollowRepo) BatchCountFollowers(ctx context.Context, repoIDs []string) (map[string]int64, error) {
	if len(repoIDs) == 0 {
		return map[string]int64{}, nil
	}

	type result struct {
		RepoID string
		Count  int64
	}
	var results []result
	err := r.db.DB.WithContext(ctx).
		Model(&schema.Follow{}).
		Select("repo_id, COUNT(*) as count").
		Where("repo_id IN ?", repoIDs).
		Group("repo_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(repoIDs))
	for _, r := range results {
		counts[r.RepoID] = r.Count
	}
	return counts, nil
}

// ListFollowedRepoIDs 获取用户关注的知识库 ID 列表.
func (r *FollowRepo) ListFollowedRepoIDs(ctx context.Context, userID string) ([]string, error) {
	var follows []schema.Follow
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&follows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.RepoID)
	}
	return ids, nil
}
