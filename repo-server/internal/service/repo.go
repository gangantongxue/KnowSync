package service

import (
	"context"
	"log/slog"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/repo-server/pkg/database/schema"
)

// CreateRepo 创建知识库，同时将创建者添加为 ADMIN 角色协作者
func (s *Service) CreateRepo(ctx context.Context, req *pb.CreateRepoRequest) (*pb.CreateRepoResponse, error) {
	repo := &schema.Repo{
		OwnerID:     req.OwnerId,
		Name:        req.Name,
		Visibility:  req.Visibility.String(),
		Description: req.Description,
	}
	if err := s.Repository.CreateRepo(ctx, repo); err != nil {
		slog.Error("创建知识库失败", "error", err)
		return &pb.CreateRepoResponse{Success: false, Msg: "创建知识库失败"}, nil
	}

	collaborator := &schema.Collaborator{
		RepoID: repo.ID,
		UserID: req.OwnerId,
		Role:   "ADMIN",
	}
	if err := s.Repository.AddCollaborator(ctx, collaborator); err != nil {
		slog.Error("添加所有者协作者失败", "error", err)
	}

	return &pb.CreateRepoResponse{
		Success: true,
		Repo:    marshalRepo(repo),
	}, nil
}

// GetRepo 获取知识库详情，同时返回当前用户的角色
func (s *Service) GetRepo(ctx context.Context, req *pb.GetRepoRequest) (*pb.GetRepoResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "VIEWER"); err != nil {
		return &pb.GetRepoResponse{Success: false, Msg: err.Error()}, nil
	}

	repo, err := s.Repository.GetRepo(ctx, req.RepoId)
	if err != nil {
		return &pb.GetRepoResponse{Success: false, Msg: "知识库不存在"}, nil
	}

	myRole := pb.CollaboratorRole_COLLABORATOR_ROLE_UNSPECIFIED
	roleStr, err := s.Repository.GetUserRole(ctx, req.RepoId, req.UserId)
	if err == nil {
		if v, ok := pb.CollaboratorRole_value[roleStr]; ok {
			myRole = pb.CollaboratorRole(v)
		}
	}

	return &pb.GetRepoResponse{
		Success: true,
		Repo:    marshalRepo(repo),
		MyRole:  myRole,
	}, nil
}

// UpdateRepo 更新知识库信息，仅 ADMIN 可操作
func (s *Service) UpdateRepo(ctx context.Context, req *pb.UpdateRepoRequest) (*pb.UpdateRepoResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "ADMIN"); err != nil {
		return &pb.UpdateRepoResponse{Success: false, Msg: err.Error()}, nil
	}

	repo, err := s.Repository.GetRepo(ctx, req.RepoId)
	if err != nil {
		return &pb.UpdateRepoResponse{Success: false, Msg: "知识库不存在"}, nil
	}

	if req.Name != "" {
		repo.Name = req.Name
	}
	if req.Description != "" {
		repo.Description = req.Description
	}
	if req.Visibility != pb.RepoVisibility_REPO_VISIBILITY_UNSPECIFIED {
		repo.Visibility = req.Visibility.String()
	}

	if err := s.Repository.UpdateRepo(ctx, repo); err != nil {
		slog.Error("更新知识库失败", "error", err)
		return &pb.UpdateRepoResponse{Success: false, Msg: "更新知识库失败"}, nil
	}

	return &pb.UpdateRepoResponse{
		Success: true,
		Repo:    marshalRepo(repo),
	}, nil
}

// DeleteRepo 软删除知识库，仅 ADMIN 可操作
func (s *Service) DeleteRepo(ctx context.Context, req *pb.DeleteRepoRequest) (*pb.DeleteRepoResponse, error) {
	if err := s.CheckRepoPermission(ctx, req.RepoId, req.UserId, "ADMIN"); err != nil {
		return &pb.DeleteRepoResponse{Success: false, Msg: err.Error()}, nil
	}

	if err := s.Repository.DeleteRepo(ctx, req.RepoId); err != nil {
		slog.Error("删除知识库失败", "error", err)
		return &pb.DeleteRepoResponse{Success: false, Msg: "删除知识库失败"}, nil
	}

	return &pb.DeleteRepoResponse{Success: true}, nil
}

// ListUserRepos 获取用户参与的所有知识库列表（先查 collaborator 获取 repo_ids，再查 repo 表）
func (s *Service) ListUserRepos(ctx context.Context, req *pb.ListUserReposRequest) (*pb.ListUserReposResponse, error) {
	repoIDs, err := s.Repository.ListUserRepoIDs(ctx, req.UserId)
	if err != nil {
		slog.Error("查询用户知识库列表失败", "error", err)
		return &pb.ListUserReposResponse{Success: false}, nil
	}

	repos, err := s.Repository.ListReposByIDs(ctx, repoIDs)
	if err != nil {
		return &pb.ListUserReposResponse{Success: false}, nil
	}

	pbRepos := make([]*pb.Repo, 0, len(repos))
	for i := range repos {
		pbRepos = append(pbRepos, marshalRepo(&repos[i]))
	}

	return &pb.ListUserReposResponse{
		Success: true,
		Repos:   pbRepos,
	}, nil
}

// marshalRepo 将数据库 Repo 转换为 protobuf Repo
func marshalRepo(r *schema.Repo) *pb.Repo {
	v, _ := pb.RepoVisibility_value[r.Visibility]
	return &pb.Repo{
		Id:           r.ID,
		OwnerId:      r.OwnerID,
		Name:         r.Name,
		Visibility:   pb.RepoVisibility(v),
		Description:  r.Description,
		ArticleCount: r.ArticleCount,
		CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
