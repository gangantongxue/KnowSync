package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// VectorizeArticle 向量化文章
// 将任务推入队列后立即返回（异步处理）
func (h *Handler) VectorizeArticle(ctx context.Context, req *pb.VectorizeArticleRequest) (*pb.VectorizeArticleResponse, error) {
	if err := h.Service.VectorizeArticle(ctx, req.GetUserId(), req.GetRepoId(), req.GetFilePath()); err != nil {
		return &pb.VectorizeArticleResponse{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}

	return &pb.VectorizeArticleResponse{
		Success: true,
		Msg:     "任务已加入队列",
	}, nil
}
