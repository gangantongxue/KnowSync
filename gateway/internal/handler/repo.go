package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CreateRepoReq 创建知识库请求体
type CreateRepoReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// UpdateRepoReq 更新知识库请求体
type UpdateRepoReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

// CreateRepo 创建知识库
func (h *Handler) CreateRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
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
		case "PUBLIC":
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

		response.Success(c, ctx, map[string]interface{}{
			"repo": map[string]interface{}{
				"id":            resp.Repo.Id,
				"name":          resp.Repo.Name,
				"visibility":    resp.Repo.Visibility.String(),
				"description":   resp.Repo.Description,
				"article_count": resp.Repo.ArticleCount,
				"created_at":    resp.Repo.CreatedAt,
			},
		})
	}
}

// GetRepo 获取知识库详情
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
			response.Error(c, ctx, 404, errcode.ErrNotFound, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]interface{}{
			"repo":    marshalRepoResponse(resp.Repo),
			"my_role": resp.MyRole.String(),
		})
	}
}

// UpdateRepo 更新知识库信息
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
		case "PUBLIC":
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

		response.Success(c, ctx, map[string]interface{}{
			"repo": marshalRepoResponse(resp.Repo),
		})
	}
}

// DeleteRepo 删除知识库
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

		response.Success(c, ctx, nil)
	}
}

// ListUserRepos 获取用户参与的所有知识库
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

		repos := make([]map[string]interface{}, 0, len(resp.Repos))
		for _, r := range resp.Repos {
			repos = append(repos, marshalRepoResponse(r))
		}

		response.Success(c, ctx, map[string]interface{}{
			"repos": repos,
		})
	}
}

// marshalRepoResponse 将 protobuf Repo 转换为 HTTP JSON 响应格式
func marshalRepoResponse(r *pb.Repo) map[string]interface{} {
	return map[string]interface{}{
		"id":            r.Id,
		"owner_id":      r.OwnerId,
		"name":          r.Name,
		"visibility":    r.Visibility.String(),
		"description":   r.Description,
		"article_count": r.ArticleCount,
		"created_at":    r.CreatedAt,
		"updated_at":    r.UpdatedAt,
	}
}

// marshalNodeResponse 将 protobuf Node 转换为 HTTP JSON 响应格式
func (h *Handler) marshalNodeResponse(n *pb.Node) map[string]interface{} {
	typ := "FOLDER"
	if n.Type == pb.NodeType_ARTICLE {
		typ = "ARTICLE"
	}
	return map[string]interface{}{
		"id":         n.Id,
		"repo_id":    n.RepoId,
		"parent_id":  n.ParentId,
		"name":       n.Name,
		"type":       typ,
		"file_path":  n.FilePath,
		"size":       n.Size,
		"created_at": n.CreatedAt,
		"updated_at": n.UpdatedAt,
	}
}

// logGrpcError 记录 gRPC 调用失败日志
func logGrpcError(service, action string, err error) {
	slog.Error("gRPC 调用失败", "service", service, "action", action, "error", err)
}
