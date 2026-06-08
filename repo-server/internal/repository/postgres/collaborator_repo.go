package postgres

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CollaboratorRepo 协作者数据访问实现.
type CollaboratorRepo struct {
	db *database.Database
}

// NewCollaboratorRepo 创建协作者数据访问实例.
func NewCollaboratorRepo(database *database.Database) *CollaboratorRepo {
	return &CollaboratorRepo{db: database}
}

// AddCollaborator 添加协作者.
func (r *CollaboratorRepo) AddCollaborator(ctx context.Context, collab *schema.Collaborator) error {
	return r.db.DB.WithContext(ctx).Create(collab).Error
}

// RemoveCollaborator 移除协作者.
func (r *CollaboratorRepo) RemoveCollaborator(ctx context.Context, repoID, userID string) error {
	return r.db.DB.WithContext(ctx).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		Delete(&schema.Collaborator{}).Error
}

// UpdateCollaborator 更新协作者角色.
func (r *CollaboratorRepo) UpdateCollaborator(ctx context.Context, repoID, userID, role string) error {
	return r.db.DB.WithContext(ctx).
		Model(&schema.Collaborator{}).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		Update("role", role).Error
}

// GetUserRole 获取用户在知识库中的角色.
func (r *CollaboratorRepo) GetUserRole(ctx context.Context, repoID, userID string) (string, error) {
	var c schema.Collaborator
	err := r.db.DB.WithContext(ctx).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		First(&c).Error
	if err != nil {
		return "", err
	}
	return c.Role, nil
}

// ListCollaborators 列出知识库的所有协作者.
func (r *CollaboratorRepo) ListCollaborators(ctx context.Context, repoID string) ([]schema.Collaborator, error) {
	var collaborators []schema.Collaborator
	err := r.db.DB.WithContext(ctx).
		Where("repo_id = ?", repoID).
		Find(&collaborators).Error
	return collaborators, err
}

// ListUserRepoIDs 获取用户有权限的知识库 ID 列表.
func (r *CollaboratorRepo) ListUserRepoIDs(ctx context.Context, userID string) ([]string, error) {
	var collaborators []schema.Collaborator
	err := r.db.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&collaborators).Error
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(collaborators))
	for _, c := range collaborators {
		ids = append(ids, c.RepoID)
	}
	return ids, nil
}
