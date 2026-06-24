package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateRepo 创建知识库.
func (h *Handler) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*pb.CreateRepoResponse, error) {
	repo, err := h.RepoService.CreateRepo(ctx, req.GetOwnerId(), req.GetName(), req.GetVisibility().String(), req.GetDescription())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateRepoResponse{
		Repo: marshalRepoWithFollowerCount(repo, 0),
	}, nil
}

// GetRepo 获取知识库详情.
func (h *Handler) GetRepo(ctx context.Context, req *pb.GetRepoRequest) (*pb.GetRepoResponse, error) {
	repo, role, err := h.RepoService.GetRepo(ctx, req.GetRepoId(), req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	myRole := pb.CollaboratorRole_COLLABORATOR_ROLE_UNSPECIFIED
	if v, ok := pb.CollaboratorRole_value[role]; ok {
		myRole = pb.CollaboratorRole(v)
	}

	followerCount, _ := h.FollowService.CountFollowers(ctx, req.GetRepoId())
	isFollowing, _ := h.FollowService.IsFollowing(ctx, req.GetRepoId(), req.GetUserId())

	return &pb.GetRepoResponse{
		Repo:        marshalRepoWithFollowerCount(repo, followerCount),
		MyRole:      myRole,
		IsFollowing: isFollowing,
	}, nil
}

// UpdateRepo 更新知识库.
func (h *Handler) UpdateRepo(ctx context.Context, req *pb.UpdateRepoRequest) (*pb.UpdateRepoResponse, error) {
	visibility := ""
	if req.GetVisibility() != pb.RepoVisibility_REPO_VISIBILITY_UNSPECIFIED {
		visibility = req.GetVisibility().String()
	}

	repo, err := h.RepoService.UpdateRepo(ctx, req.GetRepoId(), req.GetUserId(), req.GetName(), req.GetDescription(), visibility)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	followerCount, _ := h.FollowService.CountFollowers(ctx, req.GetRepoId())

	return &pb.UpdateRepoResponse{
		Repo: marshalRepoWithFollowerCount(repo, followerCount),
	}, nil
}

// DeleteRepo 删除知识库.
func (h *Handler) DeleteRepo(ctx context.Context, req *pb.DeleteRepoRequest) (*pb.DeleteRepoResponse, error) {
	if err := h.RepoService.DeleteRepo(ctx, req.GetRepoId(), req.GetUserId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteRepoResponse{}, nil
}

// ListPublicRepos 获取所有公开知识库列表.
func (h *Handler) ListPublicRepos(ctx context.Context, _ *pb.ListPublicReposRequest) (*pb.ListPublicReposResponse, error) {
	repos, err := h.RepoService.ListPublicRepos(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbRepos := h.marshalRepoListWithFollowers(ctx, repos)

	return &pb.ListPublicReposResponse{
		Repos: pbRepos,
	}, nil
}

// ListUserRepos 获取用户的知识库列表.
func (h *Handler) ListUserRepos(ctx context.Context, req *pb.ListUserReposRequest) (*pb.ListUserReposResponse, error) {
	repos, err := h.RepoService.ListUserRepos(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbRepos := h.marshalRepoListWithFollowers(ctx, repos)

	return &pb.ListUserReposResponse{
		Repos: pbRepos,
	}, nil
}

// marshalRepoListWithFollowers 批量转换 Repo 列表并填充关注数.
func (h *Handler) marshalRepoListWithFollowers(ctx context.Context, repos []schema.Repo) []*pb.Repo {
	repoIDs := make([]string, len(repos))
	for i, r := range repos {
		repoIDs[i] = r.ID
	}
	counts, _ := h.FollowService.BatchCountFollowers(ctx, repoIDs)

	pbRepos := make([]*pb.Repo, 0, len(repos))
	for _, r := range repos {
		pbRepos = append(pbRepos, marshalRepoWithFollowerCount(&r, counts[r.ID]))
	}
	return pbRepos
}

// marshalRepoWithFollowerCount 将数据库 Repo 转换为 protobuf Repo，附带关注数.
func marshalRepoWithFollowerCount(r *schema.Repo, followerCount int64) *pb.Repo {
	v := pb.RepoVisibility_value[r.Visibility]
	return &pb.Repo{
		Id:            r.ID,
		OwnerId:       r.OwnerID,
		Name:          r.Name,
		Visibility:    pb.RepoVisibility(v),
		Description:   r.Description,
		ArticleCount:  r.ArticleCount,
		CreatedAt:     r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		FollowerCount: followerCount,
	}
}

// IncrementArticleCount 原子增减知识库文章计数.
func (h *Handler) IncrementArticleCount(ctx context.Context, req *pb.IncrementArticleCountRequest) (*pb.IncrementArticleCountResponse, error) {
	if err := h.RepoService.IncrementArticleCount(ctx, req.GetRepoId(), int(req.GetDelta())); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.IncrementArticleCountResponse{}, nil
}
