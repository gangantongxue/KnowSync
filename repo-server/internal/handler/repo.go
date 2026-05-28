package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CreateRepo 创建知识库
func (h *Handler) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*pb.CreateRepoResponse, error) {
	return h.Service.CreateRepo(ctx, req)
}

// GetRepo 获取知识库详情
func (h *Handler) GetRepo(ctx context.Context, req *pb.GetRepoRequest) (*pb.GetRepoResponse, error) {
	return h.Service.GetRepo(ctx, req)
}

// UpdateRepo 更新知识库
func (h *Handler) UpdateRepo(ctx context.Context, req *pb.UpdateRepoRequest) (*pb.UpdateRepoResponse, error) {
	return h.Service.UpdateRepo(ctx, req)
}

// DeleteRepo 删除知识库
func (h *Handler) DeleteRepo(ctx context.Context, req *pb.DeleteRepoRequest) (*pb.DeleteRepoResponse, error) {
	return h.Service.DeleteRepo(ctx, req)
}

// ListUserRepos 获取用户的知识库列表
func (h *Handler) ListUserRepos(ctx context.Context, req *pb.ListUserReposRequest) (*pb.ListUserReposResponse, error) {
	return h.Service.ListUserRepos(ctx, req)
}
