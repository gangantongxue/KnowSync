package handler

import (
	"context"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// SendFriendRequest 发送好友申请
func (h *Handler) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestReq) (*pb.SendFriendRequestResp, error) {
	fr, err := h.Service.SendFriendRequest(ctx, req.GetSenderId(), req.GetReceiverId(), req.GetRemark())
	if err != nil {
		return &pb.SendFriendRequestResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.SendFriendRequestResp{
		Success: true,
		Msg:     "发送成功",
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

// GetFriendRequestsByReceiver 获取收到的好友申请
func (h *Handler) GetFriendRequestsByReceiver(ctx context.Context, req *pb.GetFriendRequestsByReceiverReq) (*pb.GetFriendRequestsResp, error) {
	requests, err := h.Service.GetFriendRequestsByReceiver(ctx, req.GetReceiverId())
	if err != nil {
		return &pb.GetFriendRequestsResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
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
		Success:        true,
		Msg:            "获取成功",
		FriendRequests: pbRequests,
	}, nil
}

// GetFriendRequestsBySender 获取发送的好友申请
func (h *Handler) GetFriendRequestsBySender(ctx context.Context, req *pb.GetFriendRequestsBySenderReq) (*pb.GetFriendRequestsResp, error) {
	requests, err := h.Service.GetFriendRequestsBySender(ctx, req.GetSenderId())
	if err != nil {
		return &pb.GetFriendRequestsResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
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
		Success:        true,
		Msg:            "获取成功",
		FriendRequests: pbRequests,
	}, nil
}

// AcceptFriendRequest 接受好友申请
func (h *Handler) AcceptFriendRequest(ctx context.Context, req *pb.AcceptFriendRequestReq) (*pb.AcceptFriendRequestResp, error) {
	if err := h.Service.AcceptFriendRequest(ctx, req.GetRequestId(), req.GetReceiverId()); err != nil {
		return &pb.AcceptFriendRequestResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.AcceptFriendRequestResp{
		Success: true,
		Msg:     "接受成功",
	}, nil
}

// RejectFriendRequest 拒绝好友申请
func (h *Handler) RejectFriendRequest(ctx context.Context, req *pb.RejectFriendRequestReq) (*pb.RejectFriendRequestResp, error) {
	if err := h.Service.RejectFriendRequest(ctx, req.GetRequestId(), req.GetReceiverId()); err != nil {
		return &pb.RejectFriendRequestResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.RejectFriendRequestResp{
		Success: true,
		Msg:     "拒绝成功",
	}, nil
}

// GetFriendList 获取好友列表
func (h *Handler) GetFriendList(ctx context.Context, req *pb.GetFriendListReq) (*pb.GetFriendListResp, error) {
	friends, err := h.Service.GetFriendList(ctx, req.GetUserId(), req.GetQuery())
	if err != nil {
		return &pb.GetFriendListResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
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
		Success: true,
		Msg:     "获取成功",
		Friends: pbFriends,
	}, nil
}

// DeleteFriend 删除好友
func (h *Handler) DeleteFriend(ctx context.Context, req *pb.DeleteFriendReq) (*pb.DeleteFriendResp, error) {
	if err := h.Service.DeleteFriend(ctx, req.GetUserId(), req.GetFriendId()); err != nil {
		return &pb.DeleteFriendResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.DeleteFriendResp{
		Success: true,
		Msg:     "删除成功",
	}, nil
}

// UpdateFriendRemark 更新好友备注
func (h *Handler) UpdateFriendRemark(ctx context.Context, req *pb.UpdateFriendRemarkReq) (*pb.UpdateFriendRemarkResp, error) {
	if err := h.Service.UpdateFriendRemark(ctx, req.GetUserId(), req.GetFriendId(), req.GetRemark()); err != nil {
		return &pb.UpdateFriendRemarkResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
	}
	return &pb.UpdateFriendRemarkResp{
		Success: true,
		Msg:     "更新成功",
	}, nil
}

// SearchUsers 搜索用户
func (h *Handler) SearchUsers(ctx context.Context, req *pb.SearchUsersReq) (*pb.SearchUsersResp, error) {
	users, err := h.Service.SearchUsers(ctx, req.GetQuery())
	if err != nil {
		return &pb.SearchUsersResp{
			Success: false,
			Msg:     err.Error(),
		}, nil
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
		Success: true,
		Msg:     "获取成功",
		Users:   pbUsers,
	}, nil
}
