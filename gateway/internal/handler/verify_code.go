package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
)

// SendVerifyCodeRequest 发送验证码请求体
type SendVerifyCodeRequest struct {
	Email string `json:"email"`
}

// ForgetPasswordRequest 忘记密码请求体
type ForgetPasswordRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	VerifyCode string `json:"verify_code"`
}

// SendVerifyCode 发送邮箱验证码
// TODO: 调用 user-server gRPC 实现
func SendVerifyCode(c context.Context, ctx *app.RequestContext) {
	var req SendVerifyCodeRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}

// ForgetPassword 忘记密码（通过邮箱验证码重置密码）
// TODO: 调用 user-server gRPC 实现
func ForgetPassword(c context.Context, ctx *app.RequestContext) {
	var req ForgetPasswordRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}
