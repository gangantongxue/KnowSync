package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// SendVerifyCodeRequest 发送验证码请求体.
type SendVerifyCodeRequest struct {
	Email string `json:"email"`
}

// ForgetPasswordRequest 忘记密码请求体.
type ForgetPasswordRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	VerifyCode string `json:"verify_code"`
}

// SendVerifyCode 发送邮箱验证码.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) SendVerifyCode() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req SendVerifyCodeRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		verifyResp, err := userClient.VerifyCode(c, &pb.VerifyCodeRequest{
			Email: req.Email,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "发送验证码失败")
			return
		}
		if !verifyResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, verifyResp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// ForgetPassword 忘记密码（通过邮箱验证码重置密码）.
func (h *Handler) ForgetPassword() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req ForgetPasswordRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		forgetResp, err := userClient.ForgetPassword(c, &pb.ForgetPasswordRequest{
			Email:      req.Email,
			Password:   req.Password,
			VerifyCode: req.VerifyCode,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "重置密码失败")
			return
		}
		if !forgetResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, forgetResp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}
