package handler

import (
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"

	"github.com/gangantongxue/knowsync/repo-server/internal/service"
)

// Handler gRPC 处理程序，实现 pb.RepoServiceServer 接口.
type Handler struct {
	pb.UnimplementedRepoServiceServer
	RepoService   *service.RepoService
	CollabService *service.CollaboratorService
	FollowService *service.FollowService
}

// NewHandler 创建 gRPC 处理程序实例.
func NewHandler(
	repoSvc *service.RepoService,
	collabSvc *service.CollaboratorService,
	followSvc *service.FollowService,
) (*Handler, error) {
	return &Handler{
		RepoService:   repoSvc,
		CollabService: collabSvc,
		FollowService: followSvc,
	}, nil
}
