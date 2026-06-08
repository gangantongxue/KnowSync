// Package handler 提供 gRPC 消息处理逻辑.
package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SendFriendRequest 发送好友申请.
func (h *Handler) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestReq) (*pb.SendFriendRequestResp, error) {
	fr, err := h.FriendService.SendFriendRequest(ctx, req.GetSenderId(), req.GetReceiverId(), req.GetRemark())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.SendFriendRequestResp{
		FriendRequest: &pb.FriendRequest{
			Id:         fr.ID,
			SenderId:   fr.SenderID,
			ReceiverId: fr.ReceiverID,
			Status:     fr.Status,
			Remark:     fr.Remark,
			CreatedAt:  fr.CreatedAt,
			UpdatedAt:  fr.UpdatedAt,
		},
	}, nil
}

// GetFriendRequestsByReceiver 获取收到的好友申请.
//
//nolint:dupl // GetFriendRequestsByReceiver/Sender 业务相似，保持独立方便理解
func (h *Handler) GetFriendRequestsByReceiver(ctx context.Context, req *pb.GetFriendRequestsByReceiverReq) (*pb.GetFriendRequestsResp, error) {
	requests, err := h.FriendService.GetFriendRequestsByReceiver(ctx, req.GetReceiverId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbRequests := make([]*pb.FriendRequest, len(requests))
	for i, r := range requests {
		pbRequests[i] = &pb.FriendRequest{
			Id:         r.ID,
			SenderId:   r.SenderID,
			ReceiverId: r.ReceiverID,
			Status:     r.Status,
			Remark:     r.Remark,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		}
	}
	return &pb.GetFriendRequestsResp{
		FriendRequests: pbRequests,
	}, nil
}

// GetFriendRequestsBySender 获取发送的好友申请.
//
//nolint:dupl // GetFriendRequestsByReceiver/Sender 业务相似，保持独立方便理解
func (h *Handler) GetFriendRequestsBySender(ctx context.Context, req *pb.GetFriendRequestsBySenderReq) (*pb.GetFriendRequestsResp, error) {
	requests, err := h.FriendService.GetFriendRequestsBySender(ctx, req.GetSenderId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbRequests := make([]*pb.FriendRequest, len(requests))
	for i, r := range requests {
		pbRequests[i] = &pb.FriendRequest{
			Id:         r.ID,
			SenderId:   r.SenderID,
			ReceiverId: r.ReceiverID,
			Status:     r.Status,
			Remark:     r.Remark,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		}
	}
	return &pb.GetFriendRequestsResp{
		FriendRequests: pbRequests,
	}, nil
}

// AcceptFriendRequest 接受好友申请.
func (h *Handler) AcceptFriendRequest(ctx context.Context, req *pb.AcceptFriendRequestReq) (*pb.AcceptFriendRequestResp, error) {
	if err := h.FriendService.AcceptFriendRequest(ctx, req.GetRequestId(), req.GetReceiverId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AcceptFriendRequestResp{}, nil
}

// RejectFriendRequest 拒绝好友申请.
func (h *Handler) RejectFriendRequest(ctx context.Context, req *pb.RejectFriendRequestReq) (*pb.RejectFriendRequestResp, error) {
	if err := h.FriendService.RejectFriendRequest(ctx, req.GetRequestId(), req.GetReceiverId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RejectFriendRequestResp{}, nil
}

// GetFriendList 获取好友列表.
func (h *Handler) GetFriendList(ctx context.Context, req *pb.GetFriendListReq) (*pb.GetFriendListResp, error) {
	friends, err := h.FriendService.GetFriendList(ctx, req.GetUserId(), req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbFriends := make([]*pb.Friend, len(friends))
	for i, f := range friends {
		pbFriends[i] = &pb.Friend{
			Id:            f.ID,
			UserId:        f.UserID,
			FriendId:      f.FriendID,
			Remark:        f.Remark,
			LastMessageAt: f.LastMessageAt,
			CreatedAt:     f.CreatedAt,
		}
	}
	return &pb.GetFriendListResp{
		Friends: pbFriends,
	}, nil
}

// DeleteFriend 删除好友.
func (h *Handler) DeleteFriend(ctx context.Context, req *pb.DeleteFriendReq) (*pb.DeleteFriendResp, error) {
	if err := h.FriendService.DeleteFriend(ctx, req.GetUserId(), req.GetFriendId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteFriendResp{}, nil
}

// UpdateFriendRemark 更新好友备注.
func (h *Handler) UpdateFriendRemark(ctx context.Context, req *pb.UpdateFriendRemarkReq) (*pb.UpdateFriendRemarkResp, error) {
	if err := h.FriendService.UpdateFriendRemark(ctx, req.GetUserId(), req.GetFriendId(), req.GetRemark()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateFriendRemarkResp{}, nil
}

// SearchUsers 搜索用户.
func (h *Handler) SearchUsers(ctx context.Context, req *pb.SearchUsersReq) (*pb.SearchUsersResp, error) {
	users, err := h.FriendService.SearchUsers(ctx, req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbUsers := make([]*pb.SearchUserInfo, len(users))
	for i, u := range users {
		pbUsers[i] = &pb.SearchUserInfo{
			Id:     u.ID,
			Name:   u.Name,
			Avatar: u.Avatar,
		}
	}
	return &pb.SearchUsersResp{
		Users: pbUsers,
	}, nil
}
