package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// GetUserProfile 获取用户首页信息（用户信息 + 公开知识库 + 好友状态）.
//
//nolint:gocyclo // 涉及并发请求和多个条件分支
func (h *Handler) GetUserProfile() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		targetUserID, err := parseUserIDParam(ctx)
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "用户 ID 格式错误")
			return
		}

		authUserID, _ := getAuthUserID(ctx)

		// 并行获取用户信息和知识库列表
		type userResult struct {
			user *pb.User
			err  error
		}
		type reposResult struct {
			repos []*pb.Repo
			err   error
		}

		userCh := make(chan userResult, 1)
		reposCh := make(chan reposResult, 1)

		// 获取用户信息
		go func() {
			conn := h.grpcClient.GetConn("user_server")
			if conn == nil {
				userCh <- userResult{err: errServiceConn}
				return
			}
			client := pb.NewUserServiceClient(conn)
			resp, err := client.GetUser(c, &pb.GetUserRequest{UserId: targetUserID})
			if err != nil {
				userCh <- userResult{err: err}
				return
			}
			if resp.User == nil {
				userCh <- userResult{err: errUserNotFound}
				return
			}
			userCh <- userResult{user: resp.User}
		}()

		// 获取用户知识库
		go func() {
			conn := h.grpcClient.GetConn("repo_server")
			if conn == nil {
				reposCh <- reposResult{err: errServiceConn}
				return
			}
			client := pb.NewRepoServiceClient(conn)
			resp, err := client.ListUserRepos(c, &pb.ListUserReposRequest{UserId: targetUserID})
			if err != nil {
				reposCh <- reposResult{err: err}
				return
			}

			reposCh <- reposResult{repos: resp.Repos}
		}()

		userResp := <-userCh
		if userResp.err != nil {
			if errors.Is(userResp.err, errUserNotFound) {
				response.Error(c, ctx, 404, errcode.ErrNotFound, "用户不存在")
				return
			}
			slog.Error("获取用户信息失败", "error", userResp.err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取用户信息失败")
			return
		}

		reposResp := <-reposCh
		if reposResp.err != nil {
			slog.Error("获取知识库列表失败", "error", reposResp.err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取知识库列表失败")
			return
		}

		// 过滤公开知识库
		publicRepos := make([]map[string]any, 0)
		for _, r := range reposResp.repos {
			if r.Visibility == pb.RepoVisibility_PUBLIC {
				publicRepos = append(publicRepos, marshalRepoResponse(r))
			}
		}

		// 确定好友状态
		friendStatus := h.determineFriendStatus(c, authUserID, targetUserID)

		response.Success(c, ctx, map[string]any{
			KeyUser: map[string]any{
				"id":      userResp.user.Id,
				KeyName:   userResp.user.Name,
				KeyAvatar: userResp.user.Avatar,
			},
			KeyRepos:        publicRepos,
			"friend_status": friendStatus,
		})
	}
}

// ListUserPublicRepos 获取用户公开知识库列表.
func (h *Handler) ListUserPublicRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		targetUserID, err := parseUserIDParam(ctx)
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "用户 ID 格式错误")
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "服务连接失败")
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListUserRepos(c, &pb.ListUserReposRequest{UserId: targetUserID})
		if err != nil {
			slog.Error("获取知识库列表失败", "error", err)
			response.Error(c, ctx, 500, errcode.ErrBadReq, "获取知识库列表失败")
			return
		}

		publicRepos := make([]map[string]any, 0)
		for _, r := range resp.Repos {
			if r.Visibility == pb.RepoVisibility_PUBLIC {
				publicRepos = append(publicRepos, marshalRepoResponse(r))
			}
		}

		response.Success(c, ctx, map[string]any{
			"repos": publicRepos,
		})
	}
}

// determineFriendStatus 判断当前登录用户与目标用户的好友关系.
//
//nolint:gocyclo // 涉及三种好友状态的多个 gRPC 查询
func (h *Handler) determineFriendStatus(c context.Context, authUserID, targetUserID string) string {
	if authUserID == "" {
		return "none"
	}
	if authUserID == targetUserID {
		return "self"
	}

	conn := h.grpcClient.GetConn("chat_server")
	if conn == nil {
		return "none"
	}

	chatClient := pb.NewChatServiceClient(conn)

	// 检查是否已是好友
	friendListResp, err := chatClient.GetFriendList(c, &pb.GetFriendListReq{UserId: authUserID})
	if err == nil {
		for _, f := range friendListResp.Friends {
			if f.FriendId == targetUserID {
				return "friends"
			}
		}
	}

	// 检查发送的好友申请
	sentResp, err := chatClient.GetFriendRequestsBySender(c, &pb.GetFriendRequestsBySenderReq{SenderId: authUserID})
	if err == nil {
		for _, req := range sentResp.FriendRequests {
			if req.ReceiverId == targetUserID {
				return "pending_sent"
			}
		}
	}

	// 检查收到的好友申请
	receivedResp, err := chatClient.GetFriendRequestsByReceiver(c, &pb.GetFriendRequestsByReceiverReq{ReceiverId: authUserID})
	if err == nil {
		for _, req := range receivedResp.FriendRequests {
			if req.SenderId == targetUserID {
				return "pending_received"
			}
		}
	}

	return "none"
}

var (
	errServiceConn    = &appError{"服务连接失败"}
	errUserNotFound   = &appError{"用户不存在"}
	errRepoListFailed = &appError{"获取知识库列表失败"}
)

type appError struct {
	msg string
}

func (e *appError) Error() string {
	return e.msg
}
