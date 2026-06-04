package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// Search 语义搜索公开知识库.
func (h *Handler) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	resp, err := h.Service.Search(ctx, req.GetQuery(), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return &pb.SearchResponse{ //nolint:nilerr // 项目约定：handler 将业务错误编码到响应体中
			Success: false,
			Msg:     err.Error(),
		}, nil
	}

	return &pb.SearchResponse{
		Success:    resp.Success,
		Msg:        resp.Msg,
		RepoIds:    resp.RepoIDs,
		TotalPages: resp.TotalPages,
		HasMore:    resp.HasMore,
	}, nil
}
