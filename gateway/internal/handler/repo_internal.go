package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// InternalListUserRepos 获取用户仓库列表（内部服务间调用）
// GET /internal/repos/user
func (h *Handler) InternalListUserRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
			ctx.JSON(consts.StatusUnauthorized, map[string]string{"message": "missing user_id"})
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListUserRepos(c, &pb.ListUserReposRequest{UserId: uid})
		if err != nil || !resp.Success {
			slog.Error("获取用户仓库列表失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "获取仓库列表失败"})
			return
		}

		repos := make([]map[string]any, 0, len(resp.Repos))
		for _, r := range resp.Repos {
			repos = append(repos, marshalInternalRepo(r))
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			"repos": repos,
			"total": len(repos),
		})
	}
}

// InternalListPublicRepos 获取公开仓库列表（内部服务间调用）
// GET /internal/repos/public
func (h *Handler) InternalListPublicRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListPublicRepos(c, &pb.ListPublicReposRequest{})
		if err != nil || !resp.Success {
			slog.Error("获取公开仓库列表失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "获取公开仓库列表失败"})
			return
		}

		repos := make([]map[string]any, 0, len(resp.Repos))
		for _, r := range resp.Repos {
			repos = append(repos, marshalInternalRepo(r))
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			"repos": repos,
			"total": len(repos),
		})
	}
}

// InternalGetRepo 获取仓库详情（内部服务间调用）
// GET /internal/repos/detail?repo_id=xxx
func (h *Handler) InternalGetRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{"message": "repo_id is required"})
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{"message": "repo server unavailable"})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.GetRepo(c, &pb.GetRepoRequest{RepoId: repoID, UserId: uid})
		if err != nil || !resp.Success {
			slog.Error("获取仓库详情失败", "error", err, "repo_id", repoID)
			ctx.JSON(consts.StatusNotFound, map[string]string{"message": "仓库不存在"})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			"repo": marshalInternalRepo(resp.Repo),
		})
	}
}

// marshalInternalRepo 将 pb.Repo 转为内部 HTTP 响应格式
func marshalInternalRepo(r *pb.Repo) map[string]any {
	return map[string]any{
		"id":            r.Id,
		"owner_id":      r.OwnerId,
		"name":          r.Name,
		"visibility":    r.Visibility.String(),
		"description":   r.Description,
		"article_count": r.ArticleCount,
	}
}
