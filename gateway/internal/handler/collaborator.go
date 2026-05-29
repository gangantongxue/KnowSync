package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CollaboratorReq 协作者操作请求体
type CollaboratorReq struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// AddCollaborator 添加协作者
func (h *Handler) AddCollaborator() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		operatorID := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		var req CollaboratorReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		role := pb.CollaboratorRole_COLLABORATOR_ROLE_UNSPECIFIED
		switch req.Role {
		case "ADMIN":
			role = pb.CollaboratorRole_ADMIN
		case "DEVELOPER":
			role = pb.CollaboratorRole_DEVELOPER
		case "VIEWER":
			role = pb.CollaboratorRole_VIEWER
		default:
			response.Error(c, ctx, 400, errcode.ErrBadReq, "角色错误，仅支持 ADMIN/DEVELOPER/VIEWER")
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.AddCollaborator(c, &pb.AddCollaboratorRequest{
			RepoId:     repoID,
			OperatorId: operatorID,
			UserId:     req.UserID,
			Role:       role,
		})
		if err != nil {
			logGrpcError("repo_server", "AddCollaborator", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "添加协作者失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// UpdateCollaborator 更新协作者角色
func (h *Handler) UpdateCollaborator() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		operatorID := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		collabUserID := ctx.Param("user_id")

		var req struct {
			Role string `json:"role"`
		}
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		role := pb.CollaboratorRole_COLLABORATOR_ROLE_UNSPECIFIED
		switch req.Role {
		case "ADMIN":
			role = pb.CollaboratorRole_ADMIN
		case "DEVELOPER":
			role = pb.CollaboratorRole_DEVELOPER
		case "VIEWER":
			role = pb.CollaboratorRole_VIEWER
		default:
			response.Error(c, ctx, 400, errcode.ErrBadReq, "角色错误，仅支持 ADMIN/DEVELOPER/VIEWER")
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.UpdateCollaborator(c, &pb.UpdateCollaboratorRequest{
			RepoId:     repoID,
			OperatorId: operatorID,
			UserId:     collabUserID,
			Role:       role,
		})
		if err != nil {
			logGrpcError("repo_server", "UpdateCollaborator", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "更新协作者失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// RemoveCollaborator 移除协作者
func (h *Handler) RemoveCollaborator() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		operatorID := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		collabUserID := ctx.Param("user_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.RemoveCollaborator(c, &pb.RemoveCollaboratorRequest{
			RepoId:     repoID,
			OperatorId: operatorID,
			UserId:     collabUserID,
		})
		if err != nil {
			logGrpcError("repo_server", "RemoveCollaborator", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "移除协作者失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// ListCollaborators 列出知识库协作者列表
func (h *Handler) ListCollaborators() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListCollaborators(c, &pb.ListCollaboratorsRequest{
			RepoId: repoID,
			UserId: uid,
		})
		if err != nil {
			logGrpcError("repo_server", "ListCollaborators", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取协作者列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "获取协作者列表失败")
			return
		}

		collaborators := make([]map[string]any, 0, len(resp.Collaborators))
		for _, c := range resp.Collaborators {
			collaborators = append(collaborators, map[string]any{
				"repo_id": c.RepoId,
				"user_id": c.UserId,
				"role":    c.Role.String(),
			})
		}

		response.Success(c, ctx, map[string]any{
			"collaborators": collaborators,
		})
	}
}
