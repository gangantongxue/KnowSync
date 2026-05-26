package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
)

// RegisterRequest 注册请求体
type RegisterRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	VerifyCode string `json:"verify_code"`
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LogoutRequest 登出请求体
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshRequest 刷新令牌请求体
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register 用户注册
// TODO: 调用 user-server gRPC 实现注册逻辑
func Register(c context.Context, ctx *app.RequestContext) {
	var req RegisterRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}

// Login 用户登录
// TODO: 调用 user-server gRPC 实现登录逻辑
func Login(c context.Context, ctx *app.RequestContext) {
	var req LoginRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}

// Logout 用户登出
// TODO: 调用 user-server gRPC 实现登出逻辑
func Logout(c context.Context, ctx *app.RequestContext) {
	var req LogoutRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}

// Refresh 刷新 access token
// TODO: 调用 user-server gRPC 实现令牌刷新逻辑
func Refresh(c context.Context, ctx *app.RequestContext) {
	var req RefreshRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}
