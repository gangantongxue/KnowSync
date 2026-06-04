package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// VerifyCode 发送邮箱验证码.
func (h *Handler) VerifyCode(ctx context.Context, req *pb.VerifyCodeRequest) (*pb.VerifyCodeResponse, error) {
	err := h.Service.VerifyCode(ctx, req.GetEmail())
	if err != nil {
		return &pb.VerifyCodeResponse{
			Success: false,
			Msg:     err.Error(),
		}, err
	}
	return &pb.VerifyCodeResponse{
		Success: true,
		Msg:     "verify code sent",
	}, nil
}

// ForgetPassword 忘记密码.
func (h *Handler) ForgetPassword(ctx context.Context, req *pb.ForgetPasswordRequest) (*pb.ForgetPasswordResponse, error) {
	err := h.Service.ForgetPassword(ctx, req.GetEmail(), req.GetPassword(), req.GetVerifyCode())
	if err != nil {
		return &pb.ForgetPasswordResponse{
			Success: false,
			Msg:     err.Error(),
		}, err
	}
	return &pb.ForgetPasswordResponse{
		Success: true,
		Msg:     "forget password success",
	}, nil
}

// ResetPassword 重置密码.
func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	err := h.Service.ResetPassword(ctx, req.GetUserId(), req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		return &pb.ResetPasswordResponse{
			Success: false,
			Msg:     err.Error(),
		}, err
	}
	return &pb.ResetPasswordResponse{
		Success: true,
		Msg:     "reset password success",
	}, nil
}
