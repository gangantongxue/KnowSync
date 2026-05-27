package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
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
func (h *Handler) GetUser() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		userID := ctx.Param("user_id")
		_ = userID
		response.Success(c, ctx, nil)
	}
}

// UpdateUser 更新用户信息
// TODO: 调用 user-server gRPC 实现
func (h *Handler) UpdateUser() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req UpdateUserRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}
		response.Success(c, ctx, nil)
	}
}

// SetAvatar 设置用户头像（multipart/form-data 文件上传）
// 流程: 校验文件魔数 → GetUser gRPC 获取旧头像 → 保存新文件 → 删除旧文件 → SetAvatar gRPC → 返回
func (h *Handler) SetAvatar() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从 JWT 上下文中获取当前登录用户 ID
		uid := ctx.GetString("user_id")
		if uid == "" {
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
		defer f.Close()

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
			// JPEG: 以 FFD8FF 开头
			ext = ".jpg"
		case n >= 4 && head[0] == 0x89 && head[1] == 0x50 && head[2] == 0x4E && head[3] == 0x47:
			// PNG: 以 89504E47 开头
			ext = ".png"
		case n >= 4 && head[0] == 0x47 && head[1] == 0x49 && head[2] == 0x46 && head[3] == 0x38:
			// GIF: 以 47494638 开头（GIF8）
			ext = ".gif"
		case n >= 12 && head[0] == 0x52 && head[1] == 0x49 && head[2] == 0x46 && head[3] == 0x46 &&
			head[8] == 0x57 && head[9] == 0x45 && head[10] == 0x42 && head[11] == 0x50:
			// WebP: RIFF 文件头 + WEBP 标识
			ext = ".webp"
		default:
			response.Error(c, ctx, 400, errcode.ErrBadReq, "不支持的头像文件格式，仅支持 JPG/PNG/GIF/WebP")
			return
		}

		// 将已读取的魔数头部与剩余文件拼回完整流
		avatarReader := io.MultiReader(bytes.NewReader(head[:n]), f)

		// --- 获取旧头像 URL，用于后续清理 ---
		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		getUserResp, err := userClient.GetUser(c, &pb.GetUserRequest{UserId: uid})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取用户信息失败")
			return
		}
		if !getUserResp.Success || getUserResp.User == nil {
			response.Error(c, ctx, 404, errcode.ErrBadReq, "用户不存在")
			return
		}

		oldAvatarURL := getUserResp.User.Avatar

		// --- 保存新头像 ---
		key := fmt.Sprintf("avatars/%s%s", uid, ext)
		if _, err := h.store.Put(storage.BucketPublic, key, avatarReader); err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "头像文件保存失败")
			return
		}

		// --- 清理旧头像文件（路径不同时）---
		if oldAvatarURL != "" {
			// 旧 URL 格式: files/public/avatars/{user_id}.{ext}
			oldKey := strings.TrimPrefix(oldAvatarURL, "files/public/")
			if oldKey != key {
				if err := h.store.Delete(storage.BucketPublic, oldKey); err != nil {
					slog.Warn("删除旧头像文件失败", "user_id", uid, "old_key", oldKey, "error", err)
				}
			}
		}

		// --- 更新数据库中的头像地址 ---
		avatarURL := fmt.Sprintf("files/public/%s", key)
		if _, err := userClient.SetAvatar(c, &pb.SetAvatarRequest{
			UserId: uid,
			Avatar: avatarURL,
		}); err != nil {
			// 文件已保存但 DB 更新失败，需通知客户端重试
			response.Error(c, ctx, 500, errcode.ErrBadReq, "设置头像失败")
			return
		}

		response.Success(c, ctx, map[string]interface{}{
			"avatar": avatarURL,
		})
	}
}

// DeleteUser 注销账户
// TODO: 调用 user-server gRPC 实现
func (h *Handler) DeleteUser() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req DeleteUserRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}
		response.Success(c, ctx, nil)
	}
}

// ResetPassword 重置密码（已登录状态）
// TODO: 调用 user-server gRPC 实现
func (h *Handler) ResetPassword() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req ResetPasswordRequest
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}
		response.Success(c, ctx, nil)
	}
}
