package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// AddCollaborator 添加协作者
func (h *Handler) AddCollaborator(ctx context.Context, req *pb.AddCollaboratorRequest) (*pb.AddCollaboratorResponse, error) {
	return h.Service.AddCollaborator(ctx, req)
}

// UpdateCollaborator 更新协作者角色
func (h *Handler) UpdateCollaborator(ctx context.Context, req *pb.UpdateCollaboratorRequest) (*pb.UpdateCollaboratorResponse, error) {
	return h.Service.UpdateCollaborator(ctx, req)
}

// RemoveCollaborator 移除协作者
func (h *Handler) RemoveCollaborator(ctx context.Context, req *pb.RemoveCollaboratorRequest) (*pb.RemoveCollaboratorResponse, error) {
	return h.Service.RemoveCollaborator(ctx, req)
}

// ListCollaborators 列出协作者列表
func (h *Handler) ListCollaborators(ctx context.Context, req *pb.ListCollaboratorsRequest) (*pb.ListCollaboratorsResponse, error) {
	return h.Service.ListCollaborators(ctx, req)
}
