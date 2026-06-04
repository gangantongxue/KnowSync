package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"gorm.io/gorm"
)

// CreateRepo 创建知识库
func (r *Repository) CreateRepo(ctx context.Context, repo *schema.Repo) error {
	return r.Database.DB.WithContext(ctx).Create(repo).Error
}

// GetRepo 根据 ID 获取知识库
func (r *Repository) GetRepo(ctx context.Context, repoID string) (*schema.Repo, error) {
	var repo schema.Repo
	err := r.Database.DB.WithContext(ctx).Where("id = ?", repoID).First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// UpdateRepo 更新知识库信息（名称、描述、可见性、文章数）
func (r *Repository) UpdateRepo(ctx context.Context, repo *schema.Repo) error {
	return r.Database.DB.WithContext(ctx).Model(&schema.Repo{}).
		Where("id = ?", repo.ID).
		Select("name", "description", "visibility", "article_count").
		Updates(repo).Error
}

// DeleteRepo 软删除知识库
func (r *Repository) DeleteRepo(ctx context.Context, repoID string) error {
	return r.Database.DB.WithContext(ctx).Where("id = ?", repoID).Delete(&schema.Repo{}).Error
}

// ListReposByIDs 根据 ID 列表批量获取知识库（替代 INNER JOIN）
func (r *Repository) ListReposByIDs(ctx context.Context, ids []string) ([]schema.Repo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var repos []schema.Repo
	err := r.Database.DB.WithContext(ctx).
		Where("id IN ?", ids).
		Order("created_at DESC").
		Find(&repos).Error
	return repos, err
}

// ListPublicRepos 获取所有公开知识库
func (r *Repository) ListPublicRepos(ctx context.Context) ([]schema.Repo, error) {
	var repos []schema.Repo
	err := r.Database.DB.WithContext(ctx).
		Where("visibility = ?", "PUBLIC").
		Order("updated_at DESC").
		Find(&repos).Error
	return repos, err
}

// IncrementArticleCount 原子增减知识库文章计数
func (r *Repository) IncrementArticleCount(ctx context.Context, repoID string, delta int) error {
	return r.Database.DB.WithContext(ctx).
		Model(&schema.Repo{}).
		Where("id = ?", repoID).
		UpdateColumn("article_count", gorm.Expr("article_count + ?", delta)).Error
}
