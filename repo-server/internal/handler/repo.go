package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CreateRepo 创建知识库
func (h *Handler) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*pb.CreateRepoResponse, error) {
	repo, err := h.Service.CreateRepo(ctx, req.GetOwnerId(), req.GetName(), req.GetVisibility().String(), req.GetDescription())
	if err != nil {
		return &pb.CreateRepoResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.CreateRepoResponse{
		Success: true,
		Repo:    marshalRepoWithFollowerCount(repo, 0),
	}, nil
}

// GetRepo 获取知识库详情
func (h *Handler) GetRepo(ctx context.Context, req *pb.GetRepoRequest) (*pb.GetRepoResponse, error) {
	repo, role, err := h.Service.GetRepo(ctx, req.GetRepoId(), req.GetUserId())
	if err != nil {
		return &pb.GetRepoResponse{Success: false, Msg: err.Error()}, nil
	}

	myRole := pb.CollaboratorRole_COLLABORATOR_ROLE_UNSPECIFIED
	if v, ok := pb.CollaboratorRole_value[role]; ok {
		myRole = pb.CollaboratorRole(v)
	}

	followerCount, _ := h.Service.Repository.CountFollowers(ctx, req.GetRepoId())
	isFollowing, _ := h.Service.Repository.IsFollowing(ctx, req.GetUserId(), req.GetRepoId())

	return &pb.GetRepoResponse{
		Success:     true,
		Repo:        marshalRepoWithFollowerCount(repo, followerCount),
		MyRole:      myRole,
		IsFollowing: isFollowing,
	}, nil
}

// UpdateRepo 更新知识库
func (h *Handler) UpdateRepo(ctx context.Context, req *pb.UpdateRepoRequest) (*pb.UpdateRepoResponse, error) {
	visibility := ""
	if req.GetVisibility() != pb.RepoVisibility_REPO_VISIBILITY_UNSPECIFIED {
		visibility = req.GetVisibility().String()
	}

	repo, err := h.Service.UpdateRepo(ctx, req.GetRepoId(), req.GetUserId(), req.GetName(), req.GetDescription(), visibility)
	if err != nil {
		return &pb.UpdateRepoResponse{Success: false, Msg: err.Error()}, nil
	}

	followerCount, _ := h.Service.Repository.CountFollowers(ctx, req.GetRepoId())

	return &pb.UpdateRepoResponse{
		Success: true,
		Repo:    marshalRepoWithFollowerCount(repo, followerCount),
	}, nil
}

// DeleteRepo 删除知识库
func (h *Handler) DeleteRepo(ctx context.Context, req *pb.DeleteRepoRequest) (*pb.DeleteRepoResponse, error) {
	if err := h.Service.DeleteRepo(ctx, req.GetRepoId(), req.GetUserId()); err != nil {
		return &pb.DeleteRepoResponse{Success: false, Msg: err.Error()}, nil
	}
	return &pb.DeleteRepoResponse{Success: true}, nil
}

// ListPublicRepos 获取所有公开知识库列表
func (h *Handler) ListPublicRepos(ctx context.Context, req *pb.ListPublicReposRequest) (*pb.ListPublicReposResponse, error) {
	repos, err := h.Service.ListPublicRepos(ctx)
	if err != nil {
		return &pb.ListPublicReposResponse{Success: false}, nil
	}

	pbRepos := h.marshalRepoListWithFollowers(ctx, repos)

	return &pb.ListPublicReposResponse{
		Success: true,
		Repos:   pbRepos,
	}, nil
}

// ListUserRepos 获取用户的知识库列表
func (h *Handler) ListUserRepos(ctx context.Context, req *pb.ListUserReposRequest) (*pb.ListUserReposResponse, error) {
	repos, err := h.Service.ListUserRepos(ctx, req.GetUserId())
	if err != nil {
		return &pb.ListUserReposResponse{Success: false}, nil
	}

	pbRepos := h.marshalRepoListWithFollowers(ctx, repos)

	return &pb.ListUserReposResponse{
		Success: true,
		Repos:   pbRepos,
	}, nil
}

// marshalRepoListWithFollowers 批量转换 Repo 列表并填充关注数
func (h *Handler) marshalRepoListWithFollowers(ctx context.Context, repos []schema.Repo) []*pb.Repo {
	repoIDs := make([]string, len(repos))
	for i, r := range repos {
		repoIDs[i] = r.ID
	}
	counts, _ := h.Service.Repository.BatchCountFollowers(ctx, repoIDs)

	pbRepos := make([]*pb.Repo, 0, len(repos))
	for _, r := range repos {
		pbRepos = append(pbRepos, marshalRepoWithFollowerCount(&r, counts[r.ID]))
	}
	return pbRepos
}

// marshalRepo 将数据库 Repo 转换为 protobuf Repo
func marshalRepo(r *schema.Repo) *pb.Repo {
	return marshalRepoWithFollowerCount(r, 0)
}

// marshalRepoWithFollowerCount 将数据库 Repo 转换为 protobuf Repo，附带关注数
func marshalRepoWithFollowerCount(r *schema.Repo, followerCount int64) *pb.Repo {
	v, _ := pb.RepoVisibility_value[r.Visibility]
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
