package handler

import (
	"context"
	"encoding/json"
	"log/slog"

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
// Deprecated: 该接口已弃用，请使用邀请流程（InviteCollaborator）
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

// InviteCollaboratorReq 邀请协作者请求体
type InviteCollaboratorReq struct {
	FriendID string `json:"friend_id"`
	RepoID   string `json:"repo_id"`
	Role     string `json:"role"`
}

// InviteCollaborator 邀请协作者（通过系统消息）
func (h *Handler) InviteCollaborator() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req InviteCollaboratorReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		// 验证角色
		switch req.Role {
		case "admin", "editor", "viewer":
		default:
			response.Error(c, ctx, 400, errcode.ErrBadReq, "角色错误，仅支持 admin/editor/viewer")
			return
		}

		repoConn := h.grpcClient.GetConn("repo_server")
		if repoConn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		// 验证仓库所有权
		repoClient := pb.NewRepoServiceClient(repoConn)
		getRepoResp, err := repoClient.GetRepo(c, &pb.GetRepoRequest{
			RepoId: req.RepoID,
			UserId: uid,
		})
		if err != nil {
			logGrpcError("repo_server", "GetRepo", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取仓库信息失败")
			return
		}
		if !getRepoResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, getRepoResp.Msg)
			return
		}
		if getRepoResp.Repo.OwnerId != uid {
			response.Error(c, ctx, 403, errcode.ErrBadReq, "只有知识库所有者可以邀请协作者")
			return
		}

		chatConn := h.grpcClient.GetConn("chat_server")
		if chatConn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		// 验证好友关系
		chatClient := pb.NewChatServiceClient(chatConn)
		friendListResp, err := chatClient.GetFriendList(c, &pb.GetFriendListReq{
			UserId: uid,
		})
		if err != nil {
			logGrpcError("chat_server", "GetFriendList", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取好友列表失败")
			return
		}
		if !friendListResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, friendListResp.Msg)
			return
		}

		isFriend := false
		for _, f := range friendListResp.Friends {
			if f.FriendId == req.FriendID {
				isFriend = true
				break
			}
		}
		if !isFriend {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "对方不是您的好友")
			return
		}

		// 发送系统邀请消息
		contentMap := map[string]string{
			"type":      "collaborator_invite",
			"repo_id":   req.RepoID,
			"repo_name": getRepoResp.Repo.Name,
			"role":      req.Role,
		}
		contentBytes, _ := json.Marshal(contentMap)

		sendResp, err := chatClient.SendPrivateMessage(c, &pb.SendPrivateMessageReq{
			SenderId:    uid,
			ReceiverId:  req.FriendID,
			ContentType: "system_invitation",
			Content:     string(contentBytes),
		})
		if err != nil {
			slog.Error("gRPC SendPrivateMessage 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "发送邀请消息失败")
			return
		}
		if !sendResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, sendResp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"message_id": sendResp.Message.Id,
		})
	}
}

// AcceptInvitation 接受协作者邀请
func (h *Handler) AcceptInvitation() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		messageID := ctx.Param("message_id")
		if messageID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "消息 ID 不能为空")
			return
		}

		chatConn := h.grpcClient.GetConn("chat_server")
		if chatConn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		// 查询消息
		chatClient := pb.NewChatServiceClient(chatConn)
		getMsgResp, err := chatClient.GetMessageByID(c, &pb.GetMessageByIDReq{
			MessageId: messageID,
		})
		if err != nil {
			slog.Error("gRPC GetMessageByID 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取消息失败")
			return
		}
		if !getMsgResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, getMsgResp.Msg)
			return
		}

		msg := getMsgResp.Message
		if msg.ContentType != "system_invitation" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "消息类型错误")
			return
		}

		// 解析消息内容
		var content struct {
			Type     string `json:"type"`
			RepoID   string `json:"repo_id"`
			RepoName string `json:"repo_name"`
			Role     string `json:"role"`
		}
		if err := json.Unmarshal([]byte(msg.Content), &content); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "消息内容解析失败")
			return
		}
		if content.Type != "collaborator_invite" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "邀请消息类型错误")
			return
		}

		// 角色映射
		role := pb.CollaboratorRole_COLLABORATOR_ROLE_UNSPECIFIED
		switch content.Role {
		case "admin":
			role = pb.CollaboratorRole_ADMIN
		case "editor":
			role = pb.CollaboratorRole_DEVELOPER
		case "viewer":
			role = pb.CollaboratorRole_VIEWER
		default:
			response.Error(c, ctx, 400, errcode.ErrBadReq, "角色错误")
			return
		}

		repoConn := h.grpcClient.GetConn("repo_server")
		if repoConn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		// 添加协作者
		repoClient := pb.NewRepoServiceClient(repoConn)
		addResp, err := repoClient.AddCollaborator(c, &pb.AddCollaboratorRequest{
			RepoId:     content.RepoID,
			OperatorId: msg.SenderId,
			UserId:     uid,
			Role:       role,
		})
		if err != nil {
			logGrpcError("repo_server", "AddCollaborator", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "添加协作者失败")
			return
		}
		if !addResp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, addResp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// RejectInvitation 拒绝协作者邀请
func (h *Handler) RejectInvitation() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		_, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		messageID := ctx.Param("message_id")
		if messageID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "消息 ID 不能为空")
			return
		}

		// 目前只返回成功，后续可更新消息状态
		response.Success(c, ctx, nil)
	}
}
