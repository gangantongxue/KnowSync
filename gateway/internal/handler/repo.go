package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CreateRepoReq 创建知识库请求体.
type CreateRepoReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// UpdateRepoReq 更新知识库请求体.
type UpdateRepoReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// CreateRepo 创建知识库.
func (h *Handler) CreateRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req CreateRepoReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		v := pb.RepoVisibility_REPO_VISIBILITY_UNSPECIFIED
		switch req.Visibility {
		case KeyVisibilityPUBLIC:
			v = pb.RepoVisibility_PUBLIC
		case "PRIVATE":
			v = pb.RepoVisibility_PRIVATE
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.CreateRepo(c, &pb.CreateRepoRequest{
			OwnerId:     uid,
			Name:        req.Name,
			Visibility:  v,
			Description: req.Description,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "创建知识库失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		if err := h.store.MakeDir(uid + "/" + resp.Repo.Id); err != nil {
			slog.Warn("创建知识库存储目录失败", "repo_id", resp.Repo.Id, "error", err)
		}

		response.Success(c, ctx, map[string]any{
			KeyRepo: map[string]any{
				"id":            resp.Repo.Id,
				KeyName:         resp.Repo.Name,
				KeyVisibility:   resp.Repo.Visibility.String(),
				KeyDescription:  resp.Repo.Description,
				KeyArticleCount: resp.Repo.ArticleCount,
				KeyCreatedAt:    resp.Repo.CreatedAt,
			},
		})
	}
}

// GetRepo 获取知识库详情.
func (h *Handler) GetRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.GetRepo(c, &pb.GetRepoRequest{
			RepoId: repoID,
			UserId: uid,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取知识库失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		if err := h.store.MakeDir(uid + "/" + resp.Repo.Id); err != nil {
			slog.Warn("创建知识库存储目录失败", "repo_id", resp.Repo.Id, "error", err)
		}

		response.Success(c, ctx, map[string]any{
			"repo":         marshalRepoResponse(resp.Repo),
			"my_role":      resp.MyRole.String(),
			"is_following": resp.IsFollowing,
		})
	}
}

// UpdateRepo 更新知识库信息.
func (h *Handler) UpdateRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		var req UpdateRepoReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		v := pb.RepoVisibility_REPO_VISIBILITY_UNSPECIFIED
		switch req.Visibility {
		case KeyVisibilityPUBLIC:
			v = pb.RepoVisibility_PUBLIC
		case "PRIVATE":
			v = pb.RepoVisibility_PRIVATE
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.UpdateRepo(c, &pb.UpdateRepoRequest{
			RepoId:      repoID,
			UserId:      uid,
			Name:        req.Name,
			Description: req.Description,
			Visibility:  v,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "更新知识库失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"repo": marshalRepoResponse(resp.Repo),
		})
	}
}

// DeleteRepo 删除知识库.
func (h *Handler) DeleteRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.DeleteRepo(c, &pb.DeleteRepoRequest{
			RepoId: repoID,
			UserId: uid,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "删除知识库失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		if err := h.store.DeleteAll(uid + "/" + repoID); err != nil {
			slog.Warn("删除知识库存储目录失败", "repo_id", repoID, "error", err)
		}

		// 异步清理向量存储，避免影响删除响应速度
		go h.deleteRepoVectors(c, repoID)

		response.Success(c, ctx, nil)
	}
}

// deleteRepoVectors 异步删除知识库向量数据.
func (h *Handler) deleteRepoVectors(c context.Context, repoID string) {
	conn := h.grpcClient.GetConn("ai_server")
	if conn == nil {
		slog.Warn("获取 ai_server 连接失败，跳过向量清理", "repo_id", repoID)
		return
	}

	client := pb.NewAIServiceClient(conn)
	_, err := client.DeleteRepoVectors(c, &pb.DeleteRepoVectorsRequest{
		RepoId: repoID,
	})
	if err != nil {
		slog.Warn("删除知识库向量数据失败", "repo_id", repoID, "error", err)
	}
}

// ListUserRepos 获取用户参与的所有知识库.
func (h *Handler) ListUserRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListUserRepos(c, &pb.ListUserReposRequest{
			UserId: uid,
		})
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取知识库列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "获取知识库列表失败")
			return
		}

		repos := make([]map[string]any, 0, len(resp.Repos))
		for _, r := range resp.Repos {
			repos = append(repos, marshalRepoResponse(r))
		}

		response.Success(c, ctx, map[string]any{
			KeyRepos: repos,
		})
	}
}

// marshalRepoResponse 将 protobuf Repo 转换为 HTTP JSON 响应格式.
func marshalRepoResponse(r *pb.Repo) map[string]any {
	return map[string]any{
		"id":             r.Id,
		KeyOwnerID:       r.OwnerId,
		KeyName:          r.Name,
		KeyVisibility:    r.Visibility.String(),
		KeyDescription:   r.Description,
		KeyArticleCount:  r.ArticleCount,
		"follower_count": r.FollowerCount,
		KeyCreatedAt:     r.CreatedAt,
		KeyUpdatedAt:     r.UpdatedAt,
	}
}

// logGrpcError 记录 gRPC 调用失败日志.
func logGrpcError(service, action string, err error) {
	slog.Error("gRPC 调用失败", "service", service, "action", action, "error", err)
}
