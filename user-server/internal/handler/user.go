package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/user-server/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserHandler 用户处理器.
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler 创建用户处理器.
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetUser 获取用户信息.
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := h.userSvc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "获取用户信息失败")
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	return &pb.GetUserResponse{
		User: &pb.User{
			Id:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Avatar: user.Avatar,
		},
	}, nil
}

// UpdateUserInfo 更新用户信息.
func (h *UserHandler) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	user, err := h.userSvc.UpdateUserInfo(ctx, req.GetUser().GetId(), req.GetUser().GetName(), req.GetUser().GetEmail(), req.GetUser().GetAvatar())
	if err != nil {
		return nil, status.Error(codes.Internal, "更新用户信息失败")
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "用户不存在")
	}

	return &pb.UpdateUserInfoResponse{
		User: &pb.User{
			Id:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Avatar: user.Avatar,
		},
	}, nil
}

// SetAvatar 设置用户头像.
func (h *UserHandler) SetAvatar(ctx context.Context, req *pb.SetAvatarRequest) (*pb.SetAvatarResponse, error) {
	err := h.userSvc.SetAvatar(ctx, req.UserId, req.Avatar)
	if err != nil {
		return nil, status.Error(codes.Internal, "设置头像失败")
	}

	return &pb.SetAvatarResponse{}, nil
}

// Unregister 注销用户.
func (h *UserHandler) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	err := h.userSvc.Unregister(ctx, req.UserId, req.Email, req.Password, req.VerifyCode)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.UnregisterResponse{}, nil
}
