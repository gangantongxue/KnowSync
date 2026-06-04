package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// SendFriendRequestReq 发送好友申请请求体.
type SendFriendRequestReq struct {
	ReceiverID string `json:"receiver_id"`
	Remark     string `json:"remark"`
}

// UpdateFriendRemarkReq 更新好友备注请求体.
type UpdateFriendRemarkReq struct {
	Remark string `json:"remark"`
}

// SendFriendRequest 发送好友申请.
func (h *Handler) SendFriendRequest() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		var req SendFriendRequestReq
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
		resp, err := client.SendFriendRequest(c, &pb.SendFriendRequestReq{
			SenderId:   uid,
			ReceiverId: req.ReceiverID,
			Remark:     req.Remark,
		})
		if err != nil {
			slog.Error("gRPC SendFriendRequest 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "发送好友申请失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, map[string]any{
			"friend_request": map[string]any{
				"id":          resp.FriendRequest.Id,
				"sender_id":   resp.FriendRequest.SenderId,
				"receiver_id": resp.FriendRequest.ReceiverId,
				"status":      resp.FriendRequest.Status,
				"remark":      resp.FriendRequest.Remark,
				KeyCreatedAt:  resp.FriendRequest.CreatedAt,
			},
		})
	}
}

// GetFriendRequestsByReceiver 获取收到的好友申请.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) GetFriendRequestsByReceiver() app.HandlerFunc {
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
		resp, err := client.GetFriendRequestsByReceiver(c, &pb.GetFriendRequestsByReceiverReq{
			ReceiverId: uid,
		})
		if err != nil {
			slog.Error("gRPC GetFriendRequestsByReceiver 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取好友申请失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		requests := make([]map[string]any, 0, len(resp.FriendRequests))
		for _, fr := range resp.FriendRequests {
			requests = append(requests, marshalFriendRequest(fr))
		}

		response.Success(c, ctx, map[string]any{
			"friend_requests": requests,
		})
	}
}

// GetFriendRequestsBySender 获取发送的好友申请.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) GetFriendRequestsBySender() app.HandlerFunc {
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
		resp, err := client.GetFriendRequestsBySender(c, &pb.GetFriendRequestsBySenderReq{
			SenderId: uid,
		})
		if err != nil {
			slog.Error("gRPC GetFriendRequestsBySender 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取好友申请失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		requests := make([]map[string]any, 0, len(resp.FriendRequests))
		for _, fr := range resp.FriendRequests {
			requests = append(requests, marshalFriendRequest(fr))
		}

		response.Success(c, ctx, map[string]any{
			"friend_requests": requests,
		})
	}
}

// AcceptFriendRequest 接受好友申请.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) AcceptFriendRequest() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		requestID := ctx.Param("request_id")
		if requestID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.AcceptFriendRequest(c, &pb.AcceptFriendRequestReq{
			RequestId:  requestID,
			ReceiverId: uid,
		})
		if err != nil {
			slog.Error("gRPC AcceptFriendRequest 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "接受好友申请失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// RejectFriendRequest 拒绝好友申请.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) RejectFriendRequest() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		requestID := ctx.Param("request_id")
		if requestID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.RejectFriendRequest(c, &pb.RejectFriendRequestReq{
			RequestId:  requestID,
			ReceiverId: uid,
		})
		if err != nil {
			slog.Error("gRPC RejectFriendRequest 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "拒绝好友申请失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// GetFriendList 获取好友列表.
func (h *Handler) GetFriendList() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		query := ctx.Query("query")

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.GetFriendList(c, &pb.GetFriendListReq{
			UserId: uid,
			Query:  query,
		})
		if err != nil {
			slog.Error("gRPC GetFriendList 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取好友列表失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		friends := make([]map[string]any, 0, len(resp.Friends))
		for _, f := range resp.Friends {
			friends = append(friends, map[string]any{
				"id":              f.Id,
				KeyUserID:         f.UserId,
				"friend_id":       f.FriendId,
				"remark":          f.Remark,
				"last_message_at": f.LastMessageAt,
				"created_at":      f.CreatedAt,
			})
		}

		response.Success(c, ctx, map[string]any{
			"friends": friends,
		})
	}
}

// DeleteFriend 删除好友.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) DeleteFriend() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		friendID := ctx.Param("friend_id")
		if friendID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "好友 ID 不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.DeleteFriend(c, &pb.DeleteFriendReq{
			UserId:   uid,
			FriendId: friendID,
		})
		if err != nil {
			slog.Error("gRPC DeleteFriend 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "删除好友失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// UpdateFriendRemark 更新好友备注.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) UpdateFriendRemark() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, ok := getAuthUserID(ctx)
		if !ok {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			return
		}

		friendID := ctx.Param("friend_id")
		if friendID == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "好友 ID 不能为空")
			return
		}

		var req UpdateFriendRemarkReq
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
		resp, err := client.UpdateFriendRemark(c, &pb.UpdateFriendRemarkReq{
			UserId:   uid,
			FriendId: friendID,
			Remark:   req.Remark,
		})
		if err != nil {
			slog.Error("gRPC UpdateFriendRemark 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "更新备注失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		response.Success(c, ctx, nil)
	}
}

// SearchUsers 搜索用户.
func (h *Handler) SearchUsers() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		query := ctx.Query("q")
		if query == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "搜索关键词不能为空")
			return
		}

		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewChatServiceClient(conn)
		resp, err := client.SearchUsers(c, &pb.SearchUsersReq{
			Query: query,
		})
		if err != nil {
			slog.Error("gRPC SearchUsers 调用失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "搜索用户失败")
			return
		}
		if !resp.Success {
			response.Error(c, ctx, 400, errcode.ErrBadReq, resp.Msg)
			return
		}

		users := make([]map[string]any, 0, len(resp.Users))
		for _, u := range resp.Users {
			users = append(users, map[string]any{
				"id":      u.Id,
				KeyName:   u.Name,
				KeyAvatar: u.Avatar,
			})
		}

		response.Success(c, ctx, map[string]any{
			"users": users,
		})
	}
}

// marshalFriendRequest 将 protobuf FriendRequest 转换为 HTTP JSON 响应格式.
func marshalFriendRequest(fr *pb.FriendRequest) map[string]any {
	return map[string]any{
		"id":          fr.Id,
		"sender_id":   fr.SenderId,
		"receiver_id": fr.ReceiverId,
		"status":      fr.Status,
		"remark":      fr.Remark,
		KeyCreatedAt:  fr.CreatedAt,
		KeyUpdatedAt:  fr.UpdatedAt,
	}
}
