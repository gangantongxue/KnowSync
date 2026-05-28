package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CreateNode 创建节点（文件夹或文章），需要 DEVELOPER 以上权限
func (s *Service) CreateNode(ctx context.Context, repoID, userID, parentID, name, nodeType string) (*schema.Node, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "ADMIN", "DEVELOPER"); err != nil {
		return nil, err
	}

	if parentID == "" {
		parentID = "__ROOT__"
	}

	count, err := s.Repository.CountNodesByName(ctx, repoID, parentID, name)
	if err != nil {
		slog.Error("检查节点名称失败", "error", err)
		return nil, fmt.Errorf("创建节点失败")
	}
	if count > 0 {
		return nil, fmt.Errorf("已存在同名节点")
	}

	node := &schema.Node{
		RepoID:   repoID,
		ParentID: parentID,
		Name:     name,
		Type:     nodeType,
	}

	if err := s.Repository.CreateNode(ctx, node); err != nil {
		slog.Error("创建节点失败", "error", err)
		return nil, fmt.Errorf("创建节点失败")
	}

	if nodeType == "ARTICLE" {
		if err := s.Repository.IncrementArticleCount(ctx, repoID, 1); err != nil {
			slog.Error("更新文章计数失败", "error", err)
		}
	}

	return node, nil
}

// GetNode 获取节点信息，需要 VIEWER 以上权限
func (s *Service) GetNode(ctx context.Context, repoID, nodeID, userID string) (*schema.Node, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "VIEWER"); err != nil {
		return nil, err
	}

	node, err := s.Repository.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}

	return node, nil
}

// UpdateNode 更新节点（重命名或移动），需要 DEVELOPER 以上权限
func (s *Service) UpdateNode(ctx context.Context, repoID, userID, nodeID, name, parentID string) (*schema.Node, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "ADMIN", "DEVELOPER"); err != nil {
		return nil, err
	}

	node, err := s.Repository.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}

	if name != "" {
		pID := node.ParentID
		if parentID != "" {
			pID = parentID
		}
		count, err := s.Repository.CountNodesByName(ctx, repoID, pID, name)
		if err != nil {
			return nil, fmt.Errorf("更新节点失败")
		}
		if count > 0 {
			return nil, fmt.Errorf("已存在同名节点")
		}
		node.Name = name
	}

	if parentID != "" {
		node.ParentID = parentID
	}

	if err := s.Repository.UpdateNode(ctx, node); err != nil {
		slog.Error("更新节点失败", "error", err)
		return nil, fmt.Errorf("更新节点失败")
	}

	return node, nil
}

// DeleteNode 递归删除节点及其所有子节点，返回所有已删除的 ID 和文件路径供清理存储
func (s *Service) DeleteNode(ctx context.Context, repoID, userID, nodeID string) ([]string, []string, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "ADMIN", "DEVELOPER"); err != nil {
		return nil, nil, err
	}

	node, err := s.Repository.GetNode(ctx, nodeID)
	if err != nil {
		return nil, nil, fmt.Errorf("节点不存在")
	}

	allIDs, err := s.Repository.GetAllDescendantIDs(ctx, nodeID)
	if err != nil {
		slog.Error("获取子节点失败", "error", err)
		return nil, nil, fmt.Errorf("删除节点失败")
	}

	filePaths, err := s.Repository.GetFilePathsByIDs(ctx, allIDs)
	if err != nil {
		slog.Error("获取文件路径失败", "error", err)
	}

	if err := s.Repository.SoftDeleteNodes(ctx, allIDs); err != nil {
		slog.Error("删除节点失败", "error", err)
		return nil, nil, fmt.Errorf("删除节点失败")
	}

	articleCount := 0
	for _, id := range allIDs {
		if id == nodeID && node.Type == "ARTICLE" {
			articleCount++
		} else {
			n, _ := s.Repository.GetNode(ctx, id)
			if n != nil && n.Type == "ARTICLE" {
				articleCount++
			}
		}
	}
	if articleCount > 0 {
		if err := s.Repository.IncrementArticleCount(ctx, repoID, -articleCount); err != nil {
			slog.Error("更新文章计数失败", "error", err)
		}
	}

	return allIDs, filePaths, nil
}

// ListNodes 列出指定父目录下的子节点列表
func (s *Service) ListNodes(ctx context.Context, repoID, userID, parentID string) ([]schema.Node, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "VIEWER"); err != nil {
		return nil, err
	}

	if parentID == "" {
		parentID = "__ROOT__"
	}

	nodes, err := s.Repository.ListNodes(ctx, repoID, parentID)
	if err != nil {
		slog.Error("查询节点列表失败", "error", err)
		return nil, fmt.Errorf("查询节点列表失败")
	}

	return nodes, nil
}

// SetArticleContent 记录文章内容在网关存储中的路径和大小
func (s *Service) SetArticleContent(ctx context.Context, repoID, userID, nodeID, filePath string, size int64) (*schema.Node, error) {
	if err := s.CheckRepoPermission(ctx, repoID, userID, "ADMIN", "DEVELOPER"); err != nil {
		return nil, err
	}

	node, err := s.Repository.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("节点不存在")
	}

	node.FilePath = filePath
	node.Size = size

	if err := s.Repository.UpdateNode(ctx, node); err != nil {
		slog.Error("更新文章内容失败", "error", err)
		return nil, fmt.Errorf("更新文章内容失败")
	}

	return node, nil
}
