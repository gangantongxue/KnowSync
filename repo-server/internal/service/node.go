package service

import (
	"context"
	"log/slog"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CreateNode 创建节点（文件夹或文章），需要 DEVELOPER 以上权限
func (s *Service) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "ADMIN", "DEVELOPER"); err != nil {
		return &pb.CreateNodeResponse{Success: false, Msg: err.Error()}, nil
	}

	parentID := req.ParentId
	if parentID == "" {
		parentID = "__ROOT__"
	}

	count, err := s.Repository.CountNodesByName(ctx, req.RepoId, parentID, req.Name)
	if err != nil {
		slog.Error("检查节点名称失败", "error", err)
		return &pb.CreateNodeResponse{Success: false, Msg: "创建节点失败"}, nil
	}
	if count > 0 {
		return &pb.CreateNodeResponse{Success: false, Msg: "已存在同名节点"}, nil
	}

	node := &schema.Node{
		RepoID:   req.RepoId,
		ParentID: parentID,
		Name:     req.Name,
		Type:     req.Type.String(),
	}

	if err := s.Repository.CreateNode(ctx, node); err != nil {
		slog.Error("创建节点失败", "error", err)
		return &pb.CreateNodeResponse{Success: false, Msg: "创建节点失败"}, nil
	}

	if req.Type == pb.NodeType_ARTICLE {
		if err := s.Repository.IncrementArticleCount(ctx, req.RepoId, 1); err != nil {
			slog.Error("更新文章计数失败", "error", err)
		}
	}

	return &pb.CreateNodeResponse{
		Success: true,
		Node:    marshalNode(node),
	}, nil
}

// GetNode 获取节点信息，需要 VIEWER 以上权限
func (s *Service) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "VIEWER"); err != nil {
		return &pb.GetNodeResponse{Success: false, Msg: err.Error()}, nil
	}

	node, err := s.Repository.GetNode(ctx, req.NodeId)
	if err != nil {
		return &pb.GetNodeResponse{Success: false, Msg: "节点不存在"}, nil
	}

	return &pb.GetNodeResponse{
		Success: true,
		Node:    marshalNode(node),
	}, nil
}

// UpdateNode 更新节点（重命名或移动），需要 DEVELOPER 以上权限
func (s *Service) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "ADMIN", "DEVELOPER"); err != nil {
		return &pb.UpdateNodeResponse{Success: false, Msg: err.Error()}, nil
	}

	node, err := s.Repository.GetNode(ctx, req.NodeId)
	if err != nil {
		return &pb.UpdateNodeResponse{Success: false, Msg: "节点不存在"}, nil
	}

	if req.Name != "" {
		parentID := node.ParentID
		if req.ParentId != "" {
			parentID = req.ParentId
		}
		count, err := s.Repository.CountNodesByName(ctx, req.RepoId, parentID, req.Name)
		if err != nil {
			return &pb.UpdateNodeResponse{Success: false, Msg: "更新节点失败"}, nil
		}
		if count > 0 {
			return &pb.UpdateNodeResponse{Success: false, Msg: "已存在同名节点"}, nil
		}
		node.Name = req.Name
	}

	if req.ParentId != "" {
		node.ParentID = req.ParentId
	}

	if err := s.Repository.UpdateNode(ctx, node); err != nil {
		slog.Error("更新节点失败", "error", err)
		return &pb.UpdateNodeResponse{Success: false, Msg: "更新节点失败"}, nil
	}

	return &pb.UpdateNodeResponse{
		Success: true,
		Node:    marshalNode(node),
	}, nil
}

// DeleteNode 递归删除节点及其所有子节点，返回所有已删除的文件路径供清理存储
func (s *Service) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "ADMIN", "DEVELOPER"); err != nil {
		return &pb.DeleteNodeResponse{Success: false, Msg: err.Error()}, nil
	}

	node, err := s.Repository.GetNode(ctx, req.NodeId)
	if err != nil {
		return &pb.DeleteNodeResponse{Success: false, Msg: "节点不存在"}, nil
	}

	allIDs, err := s.Repository.GetAllDescendantIDs(ctx, req.NodeId)
	if err != nil {
		slog.Error("获取子节点失败", "error", err)
		return &pb.DeleteNodeResponse{Success: false, Msg: "删除节点失败"}, nil
	}

	filePaths, err := s.Repository.GetFilePathsByIDs(ctx, allIDs)
	if err != nil {
		slog.Error("获取文件路径失败", "error", err)
	}

	if err := s.Repository.SoftDeleteNodes(ctx, allIDs); err != nil {
		slog.Error("删除节点失败", "error", err)
		return &pb.DeleteNodeResponse{Success: false, Msg: "删除节点失败"}, nil
	}

	articleCount := 0
	for _, id := range allIDs {
		if id == req.NodeId && node.Type == "ARTICLE" {
			articleCount++
		} else {
			n, _ := s.Repository.GetNode(ctx, id)
			if n != nil && n.Type == "ARTICLE" {
				articleCount++
			}
		}
	}
	if articleCount > 0 {
		if err := s.Repository.IncrementArticleCount(ctx, req.RepoId, -articleCount); err != nil {
			slog.Error("更新文章计数失败", "error", err)
		}
	}

	return &pb.DeleteNodeResponse{
		Success:          true,
		DeletedNodeIds:   allIDs,
		DeletedFilePaths: filePaths,
	}, nil
}

// ListNodes 列出指定父目录下的子节点列表
func (s *Service) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "VIEWER"); err != nil {
		return &pb.ListNodesResponse{Success: false, Nodes: []*pb.Node{}}, nil
	}

	parentID := req.ParentId
	if parentID == "" {
		parentID = "__ROOT__"
	}

	nodes, err := s.Repository.ListNodes(ctx, req.RepoId, parentID)
	if err != nil {
		slog.Error("查询节点列表失败", "error", err)
		return &pb.ListNodesResponse{Success: false, Nodes: []*pb.Node{}}, nil
	}

	pbNodes := make([]*pb.Node, 0, len(nodes))
	for i := range nodes {
		pbNodes = append(pbNodes, marshalNode(&nodes[i]))
	}

	return &pb.ListNodesResponse{
		Success: true,
		Nodes:   pbNodes,
	}, nil
}

// SetArticleContent 记录文章内容在网关存储中的路径和大小
func (s *Service) SetArticleContent(ctx context.Context, req *pb.SetArticleContentRequest) (*pb.SetArticleContentResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "ADMIN", "DEVELOPER"); err != nil {
		return &pb.SetArticleContentResponse{Success: false, Msg: err.Error()}, nil
	}

	node, err := s.Repository.GetNode(ctx, req.NodeId)
	if err != nil {
		return &pb.SetArticleContentResponse{Success: false, Msg: "节点不存在"}, nil
	}

	node.FilePath = req.FilePath
	node.Size = req.Size

	if err := s.Repository.UpdateNode(ctx, node); err != nil {
		slog.Error("更新文章内容失败", "error", err)
		return &pb.SetArticleContentResponse{Success: false, Msg: "更新文章内容失败"}, nil
	}

	return &pb.SetArticleContentResponse{
		Success: true,
		Node:    marshalNode(node),
	}, nil
}

// marshalNode 将数据库 Node 转换为 protobuf Node
func marshalNode(n *schema.Node) *pb.Node {
	t, _ := pb.NodeType_value[n.Type]
	return &pb.Node{
		Id:        n.ID,
		RepoId:    n.RepoID,
		ParentId:  n.ParentID,
		Name:      n.Name,
		Type:      pb.NodeType(t),
		FilePath:  n.FilePath,
		Size:      n.Size,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
