package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/gangantongxue/knowsync/chat-server/internal/repository"
	"github.com/gangantongxue/knowsync/chat-server/internal/ws"
	"github.com/gangantongxue/knowsync/chat-server/pkg/database/schema"
	"gorm.io/gorm"
)

// FriendService 好友业务服务.
type FriendService struct {
	Repo *repository.Repository
	Hub  *ws.Hub
}

// NewFriendService 创建好友服务.
func NewFriendService(repo *repository.Repository, hub *ws.Hub) *FriendService {
	return &FriendService{Repo: repo, Hub: hub}
}

// SendFriendRequest 发送好友申请.
func (s *FriendService) SendFriendRequest(ctx context.Context, senderID, receiverID, remark string) (*schema.FriendRequest, error) {
	// 检查是否是自己
	if senderID == receiverID {
		return nil, errors.New("不能添加自己为好友")
	}

	// 检查是否已经是好友
	exists, err := s.Repo.Friend.CheckFriendExists(ctx, senderID, receiverID)
	if err != nil {
		slog.Error("检查好友关系失败", "error", err)
		return nil, err
	}
	if exists {
		return nil, errors.New("已经是好友关系")
	}

	// 检查是否有待处理的好友申请
	pendingRequest, err := s.Repo.Friend.GetPendingFriendRequest(ctx, senderID, receiverID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("检查待处理好友申请失败", "error", err)
		return nil, err
	}
	if pendingRequest != nil {
		return nil, errors.New("已存在待处理的好友申请")
	}

	// 创建好友申请
	request := &schema.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     "pending",
		Remark:     remark,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	if err := s.Repo.Friend.CreateFriendRequest(ctx, request); err != nil {
		slog.Error("创建好友申请失败", "error", err)
		return nil, err
	}

	// 推送好友申请事件到接收方
	data := map[string]any{
		"request_id": request.ID,
		"sender_id":  request.SenderID,
	}
	payload, err := json.Marshal(map[string]any{
		PushKeyType: PushTypeFriendRequest,
		PushKeyData: data,
	})
	if err != nil {
		slog.Error("序列化好友申请事件失败", "error", err)
	} else {
		s.Hub.SendToUser(request.ReceiverID, payload)
	}

	return request, nil
}

// GetFriendRequestsByReceiver 获取收到的好友申请列表.
func (s *FriendService) GetFriendRequestsByReceiver(ctx context.Context, receiverID string) ([]schema.FriendRequest, error) {
	return s.Repo.Friend.GetFriendRequestsByReceiver(ctx, receiverID)
}

// GetFriendRequestsBySender 获取发送的好友申请列表.
func (s *FriendService) GetFriendRequestsBySender(ctx context.Context, senderID string) ([]schema.FriendRequest, error) {
	return s.Repo.Friend.GetFriendRequestsBySender(ctx, senderID)
}

// AcceptFriendRequest 接受好友申请.
func (s *FriendService) AcceptFriendRequest(ctx context.Context, requestID, receiverID string) error {
	// 获取好友申请
	request, err := s.Repo.Friend.GetFriendRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友申请不存在")
		}
		slog.Error("获取好友申请失败", "error", err)
		return err
	}

	// 验证接收者
	if request.ReceiverID != receiverID {
		return errors.New("无权操作此好友申请")
	}

	// 验证状态
	if request.Status != "pending" {
		return errors.New("好友申请已处理")
	}

	// 更新申请状态
	if err := s.Repo.Friend.UpdateFriendRequestStatus(ctx, requestID, "accepted"); err != nil {
		slog.Error("更新好友申请状态失败", "error", err)
		return err
	}

	// 创建双向好友关系
	now := time.Now().Unix()
	friend1 := &schema.Friend{
		UserID:        request.SenderID,
		FriendID:      request.ReceiverID,
		LastMessageAt: now,
		CreatedAt:     now,
	}
	friend2 := &schema.Friend{
		UserID:        request.ReceiverID,
		FriendID:      request.SenderID,
		LastMessageAt: now,
		CreatedAt:     now,
	}

	if err := s.Repo.Friend.CreateFriend(ctx, friend1); err != nil {
		slog.Error("创建好友关系失败", "error", err)
		return err
	}
	if err := s.Repo.Friend.CreateFriend(ctx, friend2); err != nil {
		slog.Error("创建好友关系失败", "error", err)
		return err
	}

	// 查询好友的用户信息
	var friendUser schema.User
	if err := s.Repo.DB.WithContext(ctx).Where("id = ?", receiverID).First(&friendUser).Error; err != nil {
		slog.Error("查询用户信息失败", "error", err)
	} else {
		// 推送好友接受事件到申请人
		data := map[string]any{
			"friend_id": friendUser.ID,
			"name":      friendUser.Name,
			"avatar":    friendUser.Avatar,
		}
		payload, err := json.Marshal(map[string]any{
			"type": "friend_accepted",
			"data": data,
		})
		if err != nil {
			slog.Error("序列化好友接受事件失败", "error", err)
		} else {
			s.Hub.SendToUser(request.SenderID, payload)
		}
	}

	return nil
}

// RejectFriendRequest 拒绝好友申请.
func (s *FriendService) RejectFriendRequest(ctx context.Context, requestID, receiverID string) error {
	// 获取好友申请
	request, err := s.Repo.Friend.GetFriendRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("好友申请不存在")
		}
		slog.Error("获取好友申请失败", "error", err)
		return err
	}

	// 验证接收者
	if request.ReceiverID != receiverID {
		return errors.New("无权操作此好友申请")
	}

	// 验证状态
	if request.Status != "pending" {
		return errors.New("好友申请已处理")
	}

	// 更新申请状态
	if err := s.Repo.Friend.UpdateFriendRequestStatus(ctx, requestID, "rejected"); err != nil {
		slog.Error("更新好友申请状态失败", "error", err)
		return err
	}

	return nil
}

// GetFriendList 获取好友列表.
func (s *FriendService) GetFriendList(ctx context.Context, userID, query string) ([]schema.Friend, error) {
	return s.Repo.Friend.GetFriendList(ctx, userID, query)
}

// DeleteFriend 删除好友.
func (s *FriendService) DeleteFriend(ctx context.Context, userID, friendID string) error {
	// 检查好友关系是否存在
	exists, err := s.Repo.Friend.CheckFriendExists(ctx, userID, friendID)
	if err != nil {
		slog.Error("检查好友关系失败", "error", err)
		return err
	}
	if !exists {
		return errors.New("好友关系不存在")
	}

	// 删除双向好友关系
	if err := s.Repo.Friend.DeleteFriend(ctx, userID, friendID); err != nil {
		slog.Error("删除好友关系失败", "error", err)
		return err
	}

	return nil
}

// UpdateFriendRemark 更新好友备注.
func (s *FriendService) UpdateFriendRemark(ctx context.Context, userID, friendID, remark string) error {
	// 检查好友关系是否存在
	exists, err := s.Repo.Friend.CheckFriendExists(ctx, userID, friendID)
	if err != nil {
		slog.Error("检查好友关系失败", "error", err)
		return err
	}
	if !exists {
		return errors.New("好友关系不存在")
	}

	// 更新备注
	if err := s.Repo.Friend.UpdateFriendRemark(ctx, userID, friendID, remark); err != nil {
		slog.Error("更新好友备注失败", "error", err)
		return err
	}

	return nil
}

// SearchUsers 搜索用户.
func (s *FriendService) SearchUsers(ctx context.Context, query string) ([]repository.SearchUserInfo, error) {
	return s.Repo.Friend.SearchUsers(ctx, query)
}
