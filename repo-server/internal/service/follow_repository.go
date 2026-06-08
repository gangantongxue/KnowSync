package service

import "context"

// FollowRepository 关注数据访问接口（消费者定义）.
type FollowRepository interface {
	FollowRepo(ctx context.Context, userID, repoID string) error
	UnfollowRepo(ctx context.Context, userID, repoID string) error
	IsFollowing(ctx context.Context, userID, repoID string) (bool, error)
	ListFollowedRepoIDs(ctx context.Context, userID string) ([]string, error)
	CountFollowers(ctx context.Context, repoID string) (int64, error)
	BatchCountFollowers(ctx context.Context, repoIDs []string) (map[string]int64, error)
}
