package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FollowRepo 关注知识库.
func (h *Handler) FollowRepo(ctx context.Context, req *pb.FollowRepoRequest) (*pb.FollowRepoResponse, error) {
	if err := h.FollowService.FollowRepo(ctx, req.GetRepoId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.FollowRepoResponse{}, nil
}

// UnfollowRepo 取消关注知识库.
func (h *Handler) UnfollowRepo(ctx context.Context, req *pb.UnfollowRepoRequest) (*pb.UnfollowRepoResponse, error) {
	if err := h.FollowService.UnfollowRepo(ctx, req.GetRepoId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UnfollowRepoResponse{}, nil
}

// ListFollowedRepos 获取用户关注的知识库列表.
func (h *Handler) ListFollowedRepos(ctx context.Context, req *pb.ListFollowedReposRequest) (*pb.ListFollowedReposResponse, error) {
	repos, err := h.FollowService.ListFollowedRepos(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbRepos := h.marshalRepoListWithFollowers(ctx, repos)

	return &pb.ListFollowedReposResponse{
		Repos: pbRepos,
	}, nil
}
