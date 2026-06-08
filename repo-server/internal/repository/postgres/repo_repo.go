// Package postgres 提供 repository 接口的具体数据库实现.
package postgres

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"gorm.io/gorm"
)

// RepoRepo 知识库数据访问实现.
type RepoRepo struct {
	db *database.Database
}

// NewRepoRepo 创建知识库数据访问实例.
func NewRepoRepo(database *database.Database) *RepoRepo {
	return &RepoRepo{db: database}
}

// CreateRepo 创建知识库.
func (r *RepoRepo) CreateRepo(ctx context.Context, repo *schema.Repo) error {
	return r.db.DB.WithContext(ctx).Create(repo).Error
}

// GetRepo 根据 ID 获取知识库.
func (r *RepoRepo) GetRepo(ctx context.Context, id string) (*schema.Repo, error) {
	var repo schema.Repo
	err := r.db.DB.WithContext(ctx).Where("id = ?", id).First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// UpdateRepo 更新知识库信息.
func (r *RepoRepo) UpdateRepo(ctx context.Context, repo *schema.Repo) error {
	return r.db.DB.WithContext(ctx).Model(&schema.Repo{}).
		Where("id = ?", repo.ID).
		Select("name", "description", "visibility", "article_count").
		Updates(repo).Error
}

// DeleteRepo 软删除知识库.
func (r *RepoRepo) DeleteRepo(ctx context.Context, id string) error {
	return r.db.DB.WithContext(ctx).Where("id = ?", id).Delete(&schema.Repo{}).Error
}

// ListReposByIDs 根据 ID 列表批量获取知识库.
func (r *RepoRepo) ListReposByIDs(ctx context.Context, ids []string) ([]schema.Repo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var repos []schema.Repo
	err := r.db.DB.WithContext(ctx).
		Where("id IN ?", ids).
		Order("created_at DESC").
		Find(&repos).Error
	return repos, err
}

// ListPublicRepos 获取所有公开知识库.
func (r *RepoRepo) ListPublicRepos(ctx context.Context) ([]schema.Repo, error) {
	var repos []schema.Repo
	err := r.db.DB.WithContext(ctx).
		Where("visibility = ?", "PUBLIC").
		Order("updated_at DESC").
		Find(&repos).Error
	return repos, err
}

// IncrementArticleCount 原子增减知识库文章计数.
func (r *RepoRepo) IncrementArticleCount(ctx context.Context, repoID string, delta int) error {
	return r.db.DB.WithContext(ctx).
		Model(&schema.Repo{}).
		Where("id = ?", repoID).
		UpdateColumn("article_count", gorm.Expr("article_count + ?", delta)).Error
}
