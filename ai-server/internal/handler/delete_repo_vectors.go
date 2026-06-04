package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// DeleteRepoVectors 删除知识库向量数据.
func (h *Handler) DeleteRepoVectors(ctx context.Context, req *pb.DeleteRepoVectorsRequest) (*pb.DeleteRepoVectorsResponse, error) {
	if err := h.Service.DeleteRepoVectors(ctx, req.GetRepoId()); err != nil {
		return &pb.DeleteRepoVectorsResponse{ //nolint:nilerr // 项目约定：handler 将业务错误编码到响应体中
			Success: false,
			Msg:     err.Error(),
		}, nil
	}

	return &pb.DeleteRepoVectorsResponse{
		Success: true,
	}, nil
}
