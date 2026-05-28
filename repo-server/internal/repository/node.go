package repository

import (
	"context"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CreateNode 创建节点（文件夹或文章）
func (r *Repository) CreateNode(ctx context.Context, node *schema.Node) error {
	return r.Database.DB.WithContext(ctx).Create(node).Error
}

// GetNode 根据 ID 获取节点
func (r *Repository) GetNode(ctx context.Context, nodeID string) (*schema.Node, error) {
	var node schema.Node
	err := r.Database.DB.WithContext(ctx).Where("id = ?", nodeID).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// UpdateNode 更新节点（重命名、移动位置）
func (r *Repository) UpdateNode(ctx context.Context, node *schema.Node) error {
	return r.Database.DB.WithContext(ctx).Model(&schema.Node{}).
		Where("id = ?", node.ID).
		Select("name", "parent_id", "file_path", "size").
		Updates(node).Error
}

// DeleteNode 软删除单个节点
func (r *Repository) DeleteNode(ctx context.Context, nodeID string) error {
	return r.Database.DB.WithContext(ctx).Delete(&schema.Node{}, nodeID).Error
}

// ListNodes 列出指定父节点下的所有子节点
func (r *Repository) ListNodes(ctx context.Context, repoID, parentID string) ([]schema.Node, error) {
	var nodes []schema.Node
	err := r.Database.DB.WithContext(ctx).
		Where("repo_id = ? AND parent_id = ?", repoID, parentID).
		Order("type DESC, name ASC").
		Find(&nodes).Error
	return nodes, err
}

// GetAllDescendantIDs 递归获取所有子节点 ID（使用多次查询替代 JOIN）
func (r *Repository) GetAllDescendantIDs(ctx context.Context, nodeID string) ([]string, error) {
	var ids []string
	current := []string{nodeID}

	for len(current) > 0 {
		ids = append(ids, current...)

		var children []schema.Node
		err := r.Database.DB.WithContext(ctx).
			Where("parent_id IN ?", current).
			Select("id").
			Find(&children).Error
		if err != nil {
			return nil, err
		}

		current = nil
		for _, c := range children {
			current = append(current, c.ID)
		}
	}

	return ids, nil
}

// GetFilePathsByIDs 获取指定节点列表中的文件存储路径（用于清理存储）
func (r *Repository) GetFilePathsByIDs(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var nodes []schema.Node
	err := r.Database.DB.WithContext(ctx).
		Where("id IN ? AND type = 'ARTICLE' AND file_path != ''", ids).
		Select("file_path").
		Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(nodes))
	for _, n := range nodes {
		paths = append(paths, n.FilePath)
	}
	return paths, nil
}

// SoftDeleteNodes 批量软删除节点
func (r *Repository) SoftDeleteNodes(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.Database.DB.WithContext(ctx).
		Delete(&schema.Node{}, "id IN ?", ids).Error
}

// CountNodesByName 检查同目录下是否存在同名节点（用于唯一性校验）
func (r *Repository) CountNodesByName(ctx context.Context, repoID, parentID, name string) (int64, error) {
	var count int64
	err := r.Database.DB.WithContext(ctx).
		Model(&schema.Node{}).
		Where("repo_id = ? AND parent_id = ? AND name = ?", repoID, parentID, name).
		Count(&count).Error
	return count, err
}
