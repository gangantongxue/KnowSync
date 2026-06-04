package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// DeleteFileVectors 删除指定文件的向量数据.
func (h *Handler) DeleteFileVectors(ctx context.Context, req *pb.DeleteFileVectorsRequest) (*pb.DeleteFileVectorsResponse, error) {
	if err := h.Service.DeleteFileVectors(ctx, req.GetRepoId(), req.GetFilePath()); err != nil {
		return &pb.DeleteFileVectorsResponse{ //nolint:nilerr // 项目约定：handler 将业务错误编码到响应体中
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.DeleteFileVectorsResponse{
		Success: true,
	}, nil
}
