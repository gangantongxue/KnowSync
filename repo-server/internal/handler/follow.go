package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// FollowRepo 关注知识库
func (h *Handler) FollowRepo(ctx context.Context, req *pb.FollowRepoRequest) (*pb.FollowRepoResponse, error) {
	if err := h.Service.FollowRepo(ctx, req.GetRepoId(), req.GetUserId()); err != nil {
		return &pb.FollowRepoResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.FollowRepoResponse{Success: true}, nil
}

// UnfollowRepo 取消关注知识库
func (h *Handler) UnfollowRepo(ctx context.Context, req *pb.UnfollowRepoRequest) (*pb.UnfollowRepoResponse, error) {
	if err := h.Service.UnfollowRepo(ctx, req.GetRepoId(), req.GetUserId()); err != nil {
		return &pb.UnfollowRepoResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.UnfollowRepoResponse{Success: true}, nil
}

// ListFollowedRepos 获取用户关注的知识库列表
func (h *Handler) ListFollowedRepos(ctx context.Context, req *pb.ListFollowedReposRequest) (*pb.ListFollowedReposResponse, error) {
	repos, err := h.Service.ListFollowedRepos(ctx, req.GetUserId())
	if err != nil {
		return &pb.ListFollowedReposResponse{Success: false}, nil
	}

	pbRepos := h.marshalRepoListWithFollowers(ctx, repos)

	return &pb.ListFollowedReposResponse{
		Success: true,
		Repos:   pbRepos,
	}, nil
}
