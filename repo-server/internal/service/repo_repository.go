package service

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// RepoRepository 知识库数据访问接口（消费者定义）.
type RepoRepository interface {
	CreateRepo(ctx context.Context, repo *schema.Repo) error
	GetRepo(ctx context.Context, id string) (*schema.Repo, error)
	UpdateRepo(ctx context.Context, repo *schema.Repo) error
	DeleteRepo(ctx context.Context, id string) error
	ListReposByIDs(ctx context.Context, ids []string) ([]schema.Repo, error)
	ListPublicRepos(ctx context.Context) ([]schema.Repo, error)
	IncrementArticleCount(ctx context.Context, repoID string, delta int) error
}
