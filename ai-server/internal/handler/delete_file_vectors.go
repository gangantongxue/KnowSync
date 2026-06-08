package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteFileVectors 删除指定文件的向量数据.
func (h *Handler) DeleteFileVectors(ctx context.Context, req *pb.DeleteFileVectorsRequest) (*pb.DeleteFileVectorsResponse, error) {
	if err := h.Service.DeleteFileVectors(ctx, req.GetRepoId(), req.GetFilePath()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteFileVectorsResponse{}, nil
}
