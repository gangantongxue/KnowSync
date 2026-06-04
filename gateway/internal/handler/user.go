package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// UpdateUserRequest 更新用户信息请求体.
type UpdateUserRequest struct {
	User struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar"`
	} `json:"user"`
}

// DeleteUserRequest 注销账户请求体.
type DeleteUserRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	VerifyCode string `json:"verify_code"`
}

// ResetPasswordRequest 重置密码请求体（已登录状态）.
type ResetPasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// GetUser 获取用户信息.
func (h *Handler) GetUser() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		userID, err := parseUserIDParam(ctx)
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "用户 ID 格式错误")
			return
		}

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		getUserResp, err := userClient.GetUser(c, &pb.GetUserRequest{UserId: userID})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取用户信息失败")
			return
		}
		if !getUserResp.Success || getUserResp.User == nil {
			response.Error(c, ctx, 404, errcode.ErrNotFound, "用户不存在")
			return
		}

		response.Success(c, ctx, map[string]any{
			"id":      getUserResp.User.Id,
			KeyName:   getUserResp.User.Name,
			KeyEmail:  getUserResp.User.Email,
			KeyAvatar: getUserResp.User.Avatar,
		})
	}
}

// UpdateUser 更新用户信息.
func (h *Handler) UpdateUser() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req UpdateUserRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		userID, err := parseUserIDParam(ctx)
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "用户 ID 格式错误")
			return
		}

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		updateResp, err := userClient.UpdateUserInfo(c, &pb.UpdateUserInfoRequest{
			User: &pb.User{
				Id:     userID,
				Name:   req.User.Name,
				Email:  req.User.Email,
				Avatar: req.User.Avatar,
			},
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "更新用户信息失败")
			return
		}
		if !updateResp.Success || updateResp.User == nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, updateResp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"id":     updateResp.User.Id,
			"name":   updateResp.User.Name,
			"email":  updateResp.User.Email,
			"avatar": updateResp.User.Avatar,
		})
	}
}

// SetAvatar 设置用户头像（multipart/form-data 文件上传）
// 流程: 校验文件魔数 → 保存新文件 → SetAvatar gRPC → 返回
// 使用时间戳命名文件，历史头像自动保留.
//
//nolint:gocyclo // 头像文件解析涉及多种图片格式的魔数校验
func (h *Handler) SetAvatar() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从 JWT 上下文中获取当前登录用户 ID
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		// --- 解析并校验头像文件 ---
		fileHeader, err := ctx.FormFile("avatar")
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "头像文件上传失败")
			return
		}

		f, err := fileHeader.Open()
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "头像文件读取失败")
			return
		}
		defer func() { _ = f.Close() }()

		// 读取文件头魔数，校验图片格式
		head := make([]byte, 12)
		n, readErr := io.ReadFull(f, head)
		if readErr != nil && readErr != io.ErrUnexpectedEOF {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "头像文件读取失败")
			return
		}

		var ext string
		switch {
		case n >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF:
			ext = ".jpg"
		case n >= 4 && head[0] == 0x89 && head[1] == 0x50 && head[2] == 0x4E && head[3] == 0x47:
			ext = ".png"
		case n >= 4 && head[0] == 0x47 && head[1] == 0x49 && head[2] == 0x46 && head[3] == 0x38:
			ext = ".gif"
		case n >= 12 && head[0] == 0x52 && head[1] == 0x49 && head[2] == 0x46 && head[3] == 0x46 &&
			head[8] == 0x57 && head[9] == 0x45 && head[10] == 0x42 && head[11] == 0x50:
			ext = ".webp"
		default:
			response.Error(c, ctx, 400, errcode.ErrBadReq, "不支持的头像文件格式，仅支持 JPG/PNG/GIF/WebP")
			return
		}

		// 将已读取的魔数头部与剩余文件拼回完整流
		avatarReader := io.MultiReader(bytes.NewReader(head[:n]), f)

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		// --- 保存新头像（时间戳文件名，保留历史头像）---
		ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
		key := fmt.Sprintf("%s/avatar/%s%s", uid, ts, ext)
		if _, err := h.store.WriteFile(key, avatarReader); err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "头像文件保存失败")
			return
		}

		// --- 更新数据库中的头像地址 ---
		userClient := pb.NewUserServiceClient(conn)
		avatarURL := fmt.Sprintf("/files/%s/avatar/%s%s", uid, ts, ext)
		if _, err := userClient.SetAvatar(c, &pb.SetAvatarRequest{
			UserId: uid,
			Avatar: avatarURL,
		}); err != nil {
			// 文件已保存但 DB 更新失败，需通知客户端重试
			response.Error(c, ctx, 500, errcode.ErrBadReq, "设置头像失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			"avatar": avatarURL,
		})
	}
}

// DeleteUser 注销账户.
func (h *Handler) DeleteUser() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req DeleteUserRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		userID, err := parseUserIDParam(ctx)
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "用户 ID 格式错误")
			return
		}

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		unregResp, err := userClient.Unregister(c, &pb.UnregisterRequest{
			UserId:     userID,
			Email:      req.Email,
			Password:   req.Password,
			VerifyCode: req.VerifyCode,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "注销账户失败")
			return
		}
		if !unregResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, unregResp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// ResetPassword 重置密码（已登录状态）.
func (h *Handler) ResetPassword() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req ResetPasswordRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		userID, err := parseUserIDParam(ctx)
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "用户 ID 格式错误")
			return
		}

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		resetResp, err := userClient.ResetPassword(c, &pb.ResetPasswordRequest{
			UserId:      userID,
			OldPassword: req.OldPassword,
			NewPassword: req.NewPassword,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "重置密码失败")
			return
		}
		if !resetResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resetResp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}
