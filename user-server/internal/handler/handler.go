// Package handler provides gRPC request handlers.
package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"github.com/gangantongxue/knowsync/user-server/internal/service"
)

// Handler 处理程序，实现 pb.UserServiceServer 接口.
type Handler struct {
	pb.UnimplementedUserServiceServer
	userHandler *UserHandler
	authHandler *AuthHandler
}

// NewHandler 创建处理程序.
func NewHandler(userSvc *service.UserService, authSvc *service.AuthService) *Handler {
	return &Handler{
		userHandler: NewUserHandler(userSvc),
		authHandler: NewAuthHandler(authSvc),
	}
}

// GetUser 获取用户信息.
func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return h.userHandler.GetUser(ctx, req)
}

// UpdateUserInfo 更新用户信息.
func (h *Handler) UpdateUserInfo(ctx context.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	return h.userHandler.UpdateUserInfo(ctx, req)
}

// SetAvatar 设置用户头像.
func (h *Handler) SetAvatar(ctx context.Context, req *pb.SetAvatarRequest) (*pb.SetAvatarResponse, error) {
	return h.userHandler.SetAvatar(ctx, req)
}

// Unregister 注销用户.
func (h *Handler) Unregister(ctx context.Context, req *pb.UnregisterRequest) (*pb.UnregisterResponse, error) {
	return h.userHandler.Unregister(ctx, req)
}

// Register 用户注册.
func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return h.authHandler.Register(ctx, req)
}

// Login 用户登录.
func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return h.authHandler.Login(ctx, req)
}

// Logout 用户退出登录.
func (h *Handler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return h.authHandler.Logout(ctx, req)
}

// Refresh 刷新登录凭证.
func (h *Handler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	return h.authHandler.Refresh(ctx, req)
}

// VerifyCode 发送邮箱验证码.
func (h *Handler) VerifyCode(ctx context.Context, req *pb.VerifyCodeRequest) (*pb.VerifyCodeResponse, error) {
	return h.authHandler.VerifyCode(ctx, req)
}

// ForgetPassword 忘记密码.
func (h *Handler) ForgetPassword(ctx context.Context, req *pb.ForgetPasswordRequest) (*pb.ForgetPasswordResponse, error) {
	return h.authHandler.ForgetPassword(ctx, req)
}

// ResetPassword 重置密码.
func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	return h.authHandler.ResetPassword(ctx, req)
}
