package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// Register 注册用户
func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.Service.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword(), req.GetVerifyCode())
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	if user == nil {
		return &pb.RegisterResponse{
			Success: false,
			Msg:     "user not found",
		}, nil
	}
	return &pb.RegisterResponse{
		Success: true,
		Msg:     "register success",
		User: &pb.User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

// Login 用户登录
func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	clientIP := getClientIP(ctx)
	user, accessToken, refreshToken, err := h.Service.Login(ctx, req.GetEmail(), req.GetPassword(), clientIP)
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.LoginResponse{
		Success:      true,
		Msg:          "login success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &pb.User{
			Id:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			Avatar: user.Avatar,
		},
	}, nil
}

// Logout 用户退出登录
func (h *Handler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	clientIP := getClientIP(ctx)
	err := h.Service.Logout(ctx, req.GetRefreshToken(), clientIP)
	if err != nil {
		return &pb.LogoutResponse{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.LogoutResponse{
		Success: true,
		Msg:     "logout success",
	}, nil
}

// Refresh 刷新登录凭证
func (h *Handler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	clientIP := getClientIP(ctx)
	accessToken, refreshToken, err := h.Service.Refresh(ctx, req.GetRefreshToken(), clientIP)
	if err != nil {
		return &pb.RefreshResponse{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.RefreshResponse{
		Success:      true,
		Msg:          "refresh success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
