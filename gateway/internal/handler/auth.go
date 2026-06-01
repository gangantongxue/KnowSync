package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

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

// Register 用户注册（multipart/form-data）
// 字段: name, email, password, verify_code, avatar（文件，可选）
// 流程: 解析表单 → 校验文件魔数 → Register gRPC → 保存头像 → SetAvatar gRPC → 返回
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
	var avatarReader io.Reader
	var avatarExt string

	mf, _ := ctx.MultipartForm()
	if mf != nil && mf.File != nil {
		fhs := mf.File["avatar"]
		if len(fhs) > 0 {
			fileHeader := fhs[0]
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

			// 通过文件头魔数判断真实类型，防止伪造扩展名
			switch {
			case n >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF:
				avatarExt = ".jpg"
			case n >= 4 && head[0] == 0x89 && head[1] == 0x50 && head[2] == 0x4E && head[3] == 0x47:
				avatarExt = ".png"
			case n >= 4 && head[0] == 0x47 && head[1] == 0x49 && head[2] == 0x46 && head[3] == 0x38:
				avatarExt = ".gif"
			case n >= 12 && head[0] == 0x52 && head[1] == 0x49 && head[2] == 0x46 && head[3] == 0x46 &&
				head[8] == 0x57 && head[9] == 0x45 && head[10] == 0x42 && head[11] == 0x50:
				avatarExt = ".webp"
			default:
				response.Error(c, ctx, 400, errcode.ErrBadReq, "不支持的头像文件格式，仅支持 JPG/PNG/GIF/WebP")
				return
			}

			// 将已读取的魔数头部与剩余文件拼回完整流
			avatarReader = io.MultiReader(bytes.NewReader(head[:n]), f)
		}
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

		if avatarReader != nil {
			key := fmt.Sprintf("avatars/%s%s", userID, avatarExt)
			if _, err := h.store.Put(storage.BucketPublic, key, avatarReader); err != nil {
				slog.Warn("头像文件保存失败，注册已完成但未设置头像", "user_id", userID, "error", err)
			} else {
				avatarURL = fmt.Sprintf("files/public/%s", key)
				if _, err := userClient.SetAvatar(c, &pb.SetAvatarRequest{
					UserId: userID,
					Avatar: avatarURL,
				}); err != nil {
					slog.Warn("gRPC 设置头像失败，文件已存储但数据库中未更新", "user_id", userID, "error", err)
				}
			}
		}

		response.Success(c, ctx, map[string]any{
			"user": map[string]any{
				"id":     registerResp.User.Id,
				"name":   registerResp.User.Name,
				"email":  registerResp.User.Email,
				"avatar": avatarURL,
			},
		})
	}
}

// Login 用户登录
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

		userInfo := map[string]any{}
		if loginResp.User != nil {
			userInfo = map[string]any{
				"id":     loginResp.User.Id,
				"name":   loginResp.User.Name,
				"email":  loginResp.User.Email,
				"avatar": loginResp.User.Avatar,
			}
		}
		response.Success(c, ctx, map[string]any{
			"access_token":  loginResp.AccessToken,
			"refresh_token": loginResp.RefreshToken,
			"user":          userInfo,
		})
	}
}

// Logout 用户登出
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
func (h *Handler) Refresh() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req RefreshRequest
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
		refreshResp, err := userClient.Refresh(c, &pb.RefreshRequest{
			RefreshToken: req.RefreshToken,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "刷新令牌失败")
			return
		}
		if !refreshResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, refreshResp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"access_token":  refreshResp.AccessToken,
			"refresh_token": refreshResp.RefreshToken,
		})
	}
}
