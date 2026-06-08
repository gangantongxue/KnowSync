package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteRepoVectors 删除知识库向量数据.
func (h *Handler) DeleteRepoVectors(ctx context.Context, req *pb.DeleteRepoVectorsRequest) (*pb.DeleteRepoVectorsResponse, error) {
	if err := h.Service.DeleteRepoVectors(ctx, req.GetRepoId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteRepoVectorsResponse{}, nil
}
