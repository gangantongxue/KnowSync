package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CreateNode 创建节点
func (h *Handler) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	node, err := h.Service.CreateNode(ctx, req.GetRepoId(), req.GetUserId(), req.GetParentId(), req.GetName(), req.GetType().String())
	if err != nil {
		return &pb.CreateNodeResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.CreateNodeResponse{
		Success: true,
		Node:    marshalNode(node),
	}, nil
}

// GetNode 获取节点
func (h *Handler) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	node, err := h.Service.GetNode(ctx, req.GetRepoId(), req.GetNodeId(), req.GetUserId())
	if err != nil {
		return &pb.GetNodeResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.GetNodeResponse{
		Success: true,
		Node:    marshalNode(node),
	}, nil
}

// UpdateNode 更新节点
func (h *Handler) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	node, err := h.Service.UpdateNode(ctx, req.GetRepoId(), req.GetUserId(), req.GetNodeId(), req.GetName(), req.GetParentId())
	if err != nil {
		return &pb.UpdateNodeResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.UpdateNodeResponse{
		Success: true,
		Node:    marshalNode(node),
	}, nil
}

// DeleteNode 删除节点
func (h *Handler) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeResponse, error) {
	deletedIDs, filePaths, err := h.Service.DeleteNode(ctx, req.GetRepoId(), req.GetUserId(), req.GetNodeId())
	if err != nil {
		return &pb.DeleteNodeResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.DeleteNodeResponse{
		Success:          true,
		DeletedNodeIds:   deletedIDs,
		DeletedFilePaths: filePaths,
	}, nil
}

// ListNodes 列出节点
func (h *Handler) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	nodes, err := h.Service.ListNodes(ctx, req.GetRepoId(), req.GetUserId(), req.GetParentId())
	if err != nil {
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

// SetArticleContent 记录文章存储路径
func (h *Handler) SetArticleContent(ctx context.Context, req *pb.SetArticleContentRequest) (*pb.SetArticleContentResponse, error) {
	node, err := h.Service.SetArticleContent(ctx, req.GetRepoId(), req.GetUserId(), req.GetNodeId(), req.GetFilePath(), req.GetSize())
	if err != nil {
		return &pb.SetArticleContentResponse{Success: false, Msg: err.Error()}, nil
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
