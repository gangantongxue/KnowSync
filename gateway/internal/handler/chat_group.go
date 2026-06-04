package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// CreateGroupReq 创建群组请求体
type CreateGroupReq struct {
	Name      string   `json:"name"`
	Avatar    string   `json:"avatar"`
	MemberIDs []string `json:"member_ids"`
}

// UpdateGroupReq 更新群信息请求体
type UpdateGroupReq struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// AddMembersReq 添加成员请求体
type AddMembersReq struct {
	MemberIDs []string `json:"member_ids"`
}

// TransferOwnershipReq 转让群主请求体
type TransferOwnershipReq struct {
	NewOwnerID string `json:"new_owner_id"`
}

// SetAdminReq 设置管理员请求体
type SetAdminReq struct {
	UserID string `json:"user_id"`
}

// RemoveAdminReq 移除管理员请求体
type RemoveAdminReq struct {
	UserID string `json:"user_id"`
}

// CreateGroup 创建群组
func (h *Handler) CreateGroup() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req CreateGroupReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.CreateGroup(c, &pb.CreateGroupReq{
			UserId:    uid,
			Name:      req.Name,
			Avatar:    req.Avatar,
			MemberIds: req.MemberIDs,
		})
		if err != nil {
			slog.Error("gRPC CreateGroup 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "创建群组失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"group": marshalGroup(resp.Group),
		})
	}
}

// GetGroupInfo 获取群信息
func (h *Handler) GetGroupInfo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		groupID := ctx.Param("group_id")
		if groupID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.GetGroupInfo(c, &pb.GetGroupInfoReq{
			GroupId: groupID,
			UserId:  uid,
		})
		if err != nil {
			slog.Error("gRPC GetGroupInfo 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取群信息失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"group":     marshalGroup(resp.Group),
			"is_member": resp.IsMember,
		})
	}
}

// UpdateGroup 更新群信息
func (h *Handler) UpdateGroup() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		groupID := ctx.Param("group_id")
		if groupID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 不能为空")
			return
		}

		var req UpdateGroupReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.UpdateGroup(c, &pb.UpdateGroupReq{
			GroupId: groupID,
			UserId:  uid,
			Name:    req.Name,
			Avatar:  req.Avatar,
		})
		if err != nil {
			slog.Error("gRPC UpdateGroup 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "更新群信息失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// LeaveGroup 退出群组
func (h *Handler) LeaveGroup() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		groupID := ctx.Param("group_id")
		if groupID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.LeaveGroup(c, &pb.LeaveGroupReq{
			GroupId: groupID,
			UserId:  uid,
		})
		if err != nil {
			slog.Error("gRPC LeaveGroup 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "退出群组失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// AddMembers 添加群成员
func (h *Handler) AddMembers() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		groupID := ctx.Param("group_id")
		if groupID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 不能为空")
			return
		}

		var req AddMembersReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.AddMembers(c, &pb.AddMembersReq{
			GroupId:    groupID,
			OperatorId: uid,
			MemberIds:  req.MemberIDs,
		})
		if err != nil {
			slog.Error("gRPC AddMembers 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "添加成员失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// RemoveMember 移除群成员
func (h *Handler) RemoveMember() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		groupID := ctx.Param("group_id")
		userID := ctx.Param("user_id")
		if groupID == "" || userID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 和用户 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.RemoveMember(c, &pb.RemoveMemberReq{
			GroupId:    groupID,
			OperatorId: uid,
			UserId:     userID,
		})
		if err != nil {
			slog.Error("gRPC RemoveMember 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "移除成员失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// GetGroupMembers 获取群成员列表
func (h *Handler) GetGroupMembers() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		groupID := ctx.Param("group_id")
		if groupID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.GetGroupMembers(c, &pb.GetGroupMembersReq{
			GroupId: groupID,
			UserId:  uid,
		})
		if err != nil {
			slog.Error("gRPC GetGroupMembers 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取群成员失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		members := make([]map[string]any, 0, len(resp.Members))
		for _, m := range resp.Members {
			members = append(members, map[string]any{
				"id":        m.Id,
				"group_id":  m.GroupId,
				"user_id":   m.UserId,
				"role":      m.Role,
				"joined_at": m.JoinedAt,
			})
		}

		response.Success(c, ctx, map[string]any{
			"members": members,
		})
	}
}

// TransferOwnership 转让群主
func (h *Handler) TransferOwnership() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		groupID := ctx.Param("group_id")
		if groupID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 不能为空")
			return
		}

		var req TransferOwnershipReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.TransferOwnership(c, &pb.TransferOwnershipReq{
			GroupId:        groupID,
			CurrentOwnerId: uid,
			NewOwnerId:     req.NewOwnerID,
		})
		if err != nil {
			slog.Error("gRPC TransferOwnership 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "转让群主失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// SetAdmin 设置管理员
func (h *Handler) SetAdmin() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		groupID := ctx.Param("group_id")
		if groupID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 不能为空")
			return
		}

		var req SetAdminReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.SetAdmin(c, &pb.SetAdminReq{
			GroupId:    groupID,
			OperatorId: uid,
			UserId:     req.UserID,
		})
		if err != nil {
			slog.Error("gRPC SetAdmin 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "设置管理员失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// RemoveAdmin 移除管理员
func (h *Handler) RemoveAdmin() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		groupID := ctx.Param("group_id")
		userID := ctx.Param("user_id")
		if groupID == "" || userID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "群组 ID 和用户 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.RemoveAdmin(c, &pb.RemoveAdminReq{
			GroupId:    groupID,
			OperatorId: uid,
			UserId:     userID,
		})
		if err != nil {
			slog.Error("gRPC RemoveAdmin 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "移除管理员失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// GetUserGroups 获取用户群列表
func (h *Handler) GetUserGroups() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.GetUserGroups(c, &pb.GetUserGroupsReq{
			UserId: uid,
		})
		if err != nil {
			slog.Error("gRPC GetUserGroups 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取群列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		groups := make([]map[string]any, 0, len(resp.Groups))
		for _, g := range resp.Groups {
			groups = append(groups, marshalGroup(g))
		}

		response.Success(c, ctx, map[string]any{
			"groups": groups,
		})
	}
}

// marshalGroup 将 protobuf Group 转换为 HTTP JSON 响应格式
func marshalGroup(g *pb.Group) map[string]any {
	return map[string]any{
		"id":           g.Id,
		"name":         g.Name,
		"avatar":       g.Avatar,
		"owner_id":     g.OwnerId,
		"member_count": g.MemberCount,
		"created_at":   g.CreatedAt,
		"updated_at":   g.UpdatedAt,
	}
}
