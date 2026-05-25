package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.Service.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword(), req.GetVerifyCode())
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Msg: err.Error(),
		}, err
	}
	return &pb.RegisterResponse{
		Success: true,
		Msg: "register success",
		User: &pb.User{
			Id: user.ID,
			Name: user.Name,
			Email: user.Email,
		},
	}, nil
}