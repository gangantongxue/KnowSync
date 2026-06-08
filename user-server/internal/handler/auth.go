package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/user-server/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler 认证处理器.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler 创建认证处理器.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register 用户注册.
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.authSvc.Register(ctx, req.Name, req.Email, req.Password, req.VerifyCode)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.RegisterResponse{
		User: &pb.User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

// Login 用户登录.
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	clientIP := getClientIP(ctx)
	user, accessToken, refreshToken, err := h.authSvc.Login(ctx, req.Email, req.Password, clientIP)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.LoginResponse{
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

// Logout 用户退出登录.
func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.authSvc.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "退出登录失败")
	}

	return &pb.LogoutResponse{}, nil
}

// Refresh 刷新登录凭证.
func (h *AuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	clientIP := getClientIP(ctx)
	accessToken, refreshToken, user, err := h.authSvc.Refresh(ctx, req.RefreshToken, clientIP)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.RefreshResponse{
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

// VerifyCode 发送邮箱验证码.
func (h *AuthHandler) VerifyCode(ctx context.Context, req *pb.VerifyCodeRequest) (*pb.VerifyCodeResponse, error) {
	err := h.authSvc.VerifyCode(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.VerifyCodeResponse{}, nil
}

// ForgetPassword 忘记密码.
func (h *AuthHandler) ForgetPassword(ctx context.Context, req *pb.ForgetPasswordRequest) (*pb.ForgetPasswordResponse, error) {
	err := h.authSvc.ForgetPassword(ctx, req.Email, req.Password, req.VerifyCode)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.ForgetPasswordResponse{}, nil
}

// ResetPassword 重置密码.
func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	err := h.authSvc.ResetPassword(ctx, req.UserId, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.ResetPasswordResponse{}, nil
}
