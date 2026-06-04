package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// LoginRequest 登录请求体.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LogoutRequest 登出请求体.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshRequest 刷新令牌请求体.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// setRefreshTokenCookie 设置 refresh_token 为 httpOnly cookie.
func setRefreshTokenCookie(ctx *app.RequestContext, token string) {
	// 7天过期
	maxAge := 7 * 24 * 60 * 60
	ctx.SetCookie("refresh_token", token, maxAge, "/api/v1/auth", "", protocol.CookieSameSiteLaxMode, false, true)
}

// Register 用户注册（multipart/form-data）
// 字段: name, email, password, verify_code, avatar（文件，可选）
// 流程: 解析表单 → 校验文件魔数 → Register gRPC → 保存头像 → SetAvatar gRPC → 返回.
func (h *Handler) Register() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		name := ctx.PostForm("name")
		email := ctx.PostForm("email")
		password := ctx.PostForm("password")
		verifyCode := ctx.PostForm("verify_code")

		if name == "" || email == "" || password == "" || verifyCode == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		// --- 处理可选的头像文件 ---
		avatarData, avatarExt, err := processAvatarFile(ctx)
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, err.Error())
			return
		}

		// --- 调用 Register gRPC ---
		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		registerResp, err := userClient.Register(c, &pb.RegisterRequest{
			Name:       name,
			Email:      email,
			Password:   password,
			VerifyCode: verifyCode,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "注册失败")
			return
		}
		if !registerResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, registerResp.Msg)
			return
		}

		// --- 保存头像到本地存储并通知 user-server ---
		userID := registerResp.User.Id
		avatarURL := ""

		if avatarData != nil {
			ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
			key := fmt.Sprintf("%s/avatar/%s%s", userID, ts, avatarExt)
			if _, err := h.store.WriteFile(key, bytes.NewReader(avatarData)); err != nil {
				slog.Warn("头像文件保存失败，注册已完成但未设置头像", "user_id", userID, "error", err)
			} else {
				avatarURL = fmt.Sprintf("/files/%s/avatar/%s%s", userID, ts, avatarExt)
				if _, err := userClient.SetAvatar(c, &pb.SetAvatarRequest{
					UserId: userID,
					Avatar: avatarURL,
				}); err != nil {
					slog.Warn("gRPC 设置头像失败，文件已存储但数据库中未更新", "user_id", userID, "error", err)
				}
			}
		}

		response.Success(c, ctx, map[string]any{
			KeyUser: map[string]any{
				"id":      registerResp.User.Id,
				KeyName:   registerResp.User.Name,
				KeyEmail:  registerResp.User.Email,
				KeyAvatar: avatarURL,
			},
		})
	}
}

// Login 用户登录.
func (h *Handler) Login() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req LoginRequest
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
		loginResp, err := userClient.Login(c, &pb.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "登录失败")
			return
		}
		if !loginResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, loginResp.Msg)
			return
		}

		// 设置 refresh_token 为 httpOnly cookie（安全存储）
		setRefreshTokenCookie(ctx, loginResp.RefreshToken)

		userInfo := map[string]any{}
		if loginResp.User != nil {
			userInfo = map[string]any{
				"id":      loginResp.User.Id,
				KeyName:   loginResp.User.Name,
				KeyEmail:  loginResp.User.Email,
				KeyAvatar: loginResp.User.Avatar,
			}
		}
		// 只返回 access_token（前端存储在 localStorage）
		response.Success(c, ctx, map[string]any{
			"access_token": loginResp.AccessToken,
			"user":         userInfo,
		})
	}
}

// Logout 用户登出.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) Logout() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req LogoutRequest
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
		logoutResp, err := userClient.Logout(c, &pb.LogoutRequest{
			RefreshToken: req.RefreshToken,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "登出失败")
			return
		}
		if !logoutResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, logoutResp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// Refresh 刷新 access token
// 前端 401 时自动调用，refresh_token 通过 httpOnly cookie 自动携带.
func (h *Handler) Refresh() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从 httpOnly cookie 读取 refresh_token
		refreshToken := string(ctx.Cookie("refresh_token"))
		if refreshToken == "" {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未登录")
			return
		}

		conn := h.grpcClient.GetConn("user_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		userClient := pb.NewUserServiceClient(conn)
		refreshResp, err := userClient.Refresh(c, &pb.RefreshRequest{
			RefreshToken: refreshToken,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "刷新令牌失败")
			return
		}
		if !refreshResp.Success {
			// 刷新失败，清除 cookie
			ctx.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", protocol.CookieSameSiteLaxMode, false, true)
			response.Error(c, ctx, 401, errcode.ErrUnauth, refreshResp.Msg)
			return
		}

		// 设置新的 refresh_token cookie
		setRefreshTokenCookie(ctx, refreshResp.RefreshToken)

		// 返回新的 access_token 和用户信息
		response.Success(c, ctx, map[string]any{
			"access_token": refreshResp.AccessToken,
			KeyUser: map[string]any{
				"id":      refreshResp.User.Id,
				KeyName:   refreshResp.User.Name,
				KeyEmail:  refreshResp.User.Email,
				KeyAvatar: refreshResp.User.Avatar,
			},
		})
	}
}

// processAvatarFile 从 multipart 表单中解析头像文件，返回完整内容和扩展名.
//
//nolint:gocyclo // 文件魔数校验涉及多种图片格式分支
func processAvatarFile(ctx *app.RequestContext) ([]byte, string, error) {
	mf, _ := ctx.MultipartForm()
	if mf == nil || mf.File == nil {
		return nil, "", nil
	}
	fhs := mf.File["avatar"]
	if len(fhs) == 0 {
		return nil, "", nil
	}

	fileHeader := fhs[0]
	f, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = f.Close() }()

	head := make([]byte, 12)
	n, readErr := io.ReadFull(f, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		return nil, "", readErr
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
		return nil, "", errors.New("不支持的头像文件格式，仅支持 JPG/PNG/GIF/WebP")
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, "", err
	}

	return append(head[:n], data...), ext, nil
}
