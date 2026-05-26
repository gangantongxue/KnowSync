package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// GetUser 获取用户信息
func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// TODO: implement me
	user, err := h.Service.GetUser(ctx, req.GetUserId())
	if err != nil {
		return &pb.GetUserResponse{
			Success: false,
			Msg:     err.Error(),
		}, err
	}
	if user == nil {
		return &pb.GetUserResponse{
			Success: false,
			Msg:     "user not found",
		}, nil
	}
	return &pb.GetUserResponse{
		Success: true,
		Msg:     "get user success",
		User: &pb.User{
			Id:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Avatar: user.Avatar,
		},
	}, nil
}

// UpdateUserInfo 更新用户信息
func (h *Handler) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	// TODO: implement me
	user, err := h.Service.UpdateUserInfo(ctx, req.GetUser().GetId(), req.GetUser().GetName(), req.GetUser().GetEmail(), req.GetUser().GetAvatar())
	if err != nil {
		return &pb.UpdateUserInfoResponse{
			Success: false,
			Msg:     err.Error(),
		}, err
	}
	if user == nil {
		return &pb.UpdateUserInfoResponse{
			Success: false,
			Msg:     "user not found",
		}, nil
	}
	return &pb.UpdateUserInfoResponse{
		Success: true,
		Msg:     "update user info success",
		User: &pb.User{
			Id:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Avatar: user.Avatar,
		},
	}, nil
}

// SetAvatar 设置用户头像
func (h *Handler) SetAvatar(ctx context.Context, req *pb.SetAvatarRequest) (*pb.SetAvatarResponse, error) {
	// TODO: implement me
	err := h.Service.SetAvatar(ctx, req.GetUserId(), req.GetAvatar())
	if err != nil {
		return &pb.SetAvatarResponse{
			Success: false,
			Msg:     err.Error(),
		}, err
	}
	return &pb.SetAvatarResponse{
		Success: true,
		Msg:     "set avatar success",
	}, nil
}

// Unregister 注销用户
func (h *Handler) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	// TODO: implement me
	err := h.Service.Unregister(ctx, req.GetUserId(), req.GetEmail(), req.GetPassword(), req.GetVerifyCode())
	if err != nil {
		return &pb.UnregisterResponse{
			Success: false,
			Msg:     err.Error(),
		}, err
	}
	return &pb.UnregisterResponse{
		Success: true,
		Msg:     "unregister success",
	}, nil
}
