package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Search 语义搜索公开知识库.
func (h *Handler) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	token := extractServiceToken(ctx)
	ctx = withServiceToken(ctx, token)

	resp, err := h.Service.Search(ctx, req.GetQuery(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SearchResponse{
		RepoIds:    resp.RepoIDs,
		TotalPages: resp.TotalPages,
		HasMore:    resp.HasMore,
	}, nil
}
