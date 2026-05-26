package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
)

// UpdateUserRequest 更新用户信息请求体
type UpdateUserRequest struct {
	User struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar"`
	} `json:"user"`
}

// DeleteUserRequest 注销账户请求体
type DeleteUserRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	VerifyCode string `json:"verify_code"`
}

// ResetPasswordRequest 重置密码请求体（已登录状态）
type ResetPasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// GetUser 获取用户信息
// TODO: 调用 user-server gRPC 实现
func GetUser(c context.Context, ctx *app.RequestContext) {
	userID := ctx.Param("user_id")
	_ = userID
	response.Success(c, ctx, nil)
}

// UpdateUser 更新用户信息
// TODO: 调用 user-server gRPC 实现
func UpdateUser(c context.Context, ctx *app.RequestContext) {
	var req UpdateUserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}

// SetAvatar 设置用户头像（multipart/form-data 文件上传）
// TODO: 调用 user-server gRPC 实现
func SetAvatar(c context.Context, ctx *app.RequestContext) {
	response.Success(c, ctx, nil)
}

// DeleteUser 注销账户
// TODO: 调用 user-server gRPC 实现
func DeleteUser(c context.Context, ctx *app.RequestContext) {
	var req DeleteUserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}

// ResetPassword 重置密码（已登录状态）
// TODO: 调用 user-server gRPC 实现
func ResetPassword(c context.Context, ctx *app.RequestContext) {
	var req ResetPasswordRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(c, ctx, 400, 40000, "请求参数错误")
		return
	}
	response.Success(c, ctx, nil)
}
