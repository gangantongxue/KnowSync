package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CreateNode 创建节点
func (h *Handler) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeResponse, error) {
	return h.Service.CreateNode(ctx, req)
}

// GetNode 获取节点
func (h *Handler) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	return h.Service.GetNode(ctx, req)
}

// UpdateNode 更新节点
func (h *Handler) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	return h.Service.UpdateNode(ctx, req)
}

// DeleteNode 删除节点
func (h *Handler) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeResponse, error) {
	return h.Service.DeleteNode(ctx, req)
}

// ListNodes 列出节点
func (h *Handler) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	return h.Service.ListNodes(ctx, req)
}

// SetArticleContent 记录文章存储路径
func (h *Handler) SetArticleContent(ctx context.Context, req *pb.SetArticleContentRequest) (*pb.SetArticleContentResponse, error) {
	return h.Service.SetArticleContent(ctx, req)
}
