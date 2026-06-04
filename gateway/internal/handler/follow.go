package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// FollowRepo 关注知识库.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) FollowRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.FollowRepo(c, &pb.FollowRepoRequest{
			RepoId: repoID,
			UserId: uid,
		})
		if err != nil {
			logGrpcError("repo_server", "FollowRepo", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "关注知识库失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// UnfollowRepo 取消关注知识库.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) UnfollowRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.UnfollowRepo(c, &pb.UnfollowRepoRequest{
			RepoId: repoID,
			UserId: uid,
		})
		if err != nil {
			logGrpcError("repo_server", "UnfollowRepo", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "取消关注失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// ListFollowedRepos 获取用户关注的知识库列表.
func (h *Handler) ListFollowedRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListFollowedRepos(c, &pb.ListFollowedReposRequest{
			UserId: uid,
		})
		if err != nil {
			logGrpcError("repo_server", "ListFollowedRepos", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取关注列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "获取关注列表失败")
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
