// Package repository 提供数据访问层实现，封装所有数据库操作.
package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// AddCollaborator 添加协作者.
func (r *Repository) AddCollaborator(ctx context.Context, c *schema.Collaborator) error {
	return r.Database.DB.WithContext(ctx).Create(c).Error
}

// UpdateCollaborator 更新协作者角色.
func (r *Repository) UpdateCollaborator(ctx context.Context, repoID, userID, role string) error {
	return r.Database.DB.WithContext(ctx).
		Model(&schema.Collaborator{}).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		Update("role", role).Error
}

// RemoveCollaborator 移除协作者.
func (r *Repository) RemoveCollaborator(ctx context.Context, repoID, userID string) error {
	return r.Database.DB.WithContext(ctx).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		Delete(&schema.Collaborator{}).Error
}

// ListCollaborators 列出知识库的所有协作者.
func (r *Repository) ListCollaborators(ctx context.Context, repoID string) ([]schema.Collaborator, error) {
	var collaborators []schema.Collaborator
	err := r.Database.DB.WithContext(ctx).
		Where("repo_id = ?", repoID).
		Find(&collaborators).Error
	return collaborators, err
}

// GetUserRole 获取用户在知识库中的角色.
func (r *Repository) GetUserRole(ctx context.Context, repoID, userID string) (string, error) {
	var c schema.Collaborator
	err := r.Database.DB.WithContext(ctx).
		Where("repo_id = ? AND user_id = ?", repoID, userID).
		First(&c).Error
	if err != nil {
		return "", err
	}
	return c.Role, nil
}

// ListUserRepoIDs 获取用户有权限的知识库 ID 列表（替代 JOIN 查询）.
func (r *Repository) ListUserRepoIDs(ctx context.Context, userID string) ([]string, error) {
	var collaborators []schema.Collaborator
	err := r.Database.DB.WithContext(ctx).
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
