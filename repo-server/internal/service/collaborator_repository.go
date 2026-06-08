package service

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CollaboratorRepository 协作者数据访问接口（消费者定义）.
type CollaboratorRepository interface {
	AddCollaborator(ctx context.Context, collab *schema.Collaborator) error
	RemoveCollaborator(ctx context.Context, repoID, userID string) error
	UpdateCollaborator(ctx context.Context, repoID, userID, role string) error
	GetUserRole(ctx context.Context, repoID, userID string) (string, error)
	ListCollaborators(ctx context.Context, repoID string) ([]schema.Collaborator, error)
	ListUserRepoIDs(ctx context.Context, userID string) ([]string, error)
}
