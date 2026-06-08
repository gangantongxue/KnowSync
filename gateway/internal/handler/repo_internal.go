package handler

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// InternalListUserRepos 获取用户仓库列表（内部服务间调用）
// GET /internal/repos/user.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) InternalListUserRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
			ctx.JSON(consts.StatusUnauthorized, map[string]string{KeyMessage: ErrMsgMissingUserID})
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListUserRepos(c, &pb.ListUserReposRequest{UserId: uid})
		if err != nil {
			slog.Error("获取用户仓库列表失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "获取仓库列表失败"})
			return
		}

		repos := make([]map[string]any, 0, len(resp.Repos))
		for _, r := range resp.Repos {
			repos = append(repos, marshalInternalRepo(r))
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			KeyRepos: repos,
			"total":  len(repos),
		})
	}
}

// InternalListPublicRepos 获取公开仓库列表（内部服务间调用）
// GET /internal/repos/public.
func (h *Handler) InternalListPublicRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListPublicRepos(c, &pb.ListPublicReposRequest{})
		if err != nil {
			slog.Error("获取公开仓库列表失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "获取公开仓库列表失败"})
			return
		}

		repos := make([]map[string]any, 0, len(resp.Repos))
		for _, r := range resp.Repos {
			repos = append(repos, marshalInternalRepo(r))
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			KeyRepos: repos,
			"total":  len(repos),
		})
	}
}

// InternalGetRepo 获取仓库详情（内部服务间调用）
// GET /internal/repos/detail?repo_id=xxx.
func (h *Handler) InternalGetRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgRepoIDRequired})
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.GetRepo(c, &pb.GetRepoRequest{RepoId: repoID, UserId: uid})
		if err != nil {
			slog.Error("获取仓库详情失败", "error", err, "repo_id", repoID)
			ctx.JSON(consts.StatusNotFound, map[string]string{KeyMessage: "仓库不存在"})
			return
		}

		ctx.JSON(consts.StatusOK, marshalInternalRepoDetail(resp))
	}
}

// marshalInternalRepo 将 pb.Repo 转为内部 HTTP 响应格式.
func marshalInternalRepo(r *pb.Repo) map[string]any {
	return map[string]any{
		"id":             r.Id,
		KeyOwnerID:       r.OwnerId,
		KeyName:          r.Name,
		KeyVisibility:    r.Visibility.String(),
		KeyDescription:   r.Description,
		KeyArticleCount:  r.ArticleCount,
		"follower_count": r.FollowerCount,
	}
}

// marshalInternalRepoDetail 将 pb.GetRepoResponse 转为内部 HTTP 响应格式（含角色和关注状态）.
func marshalInternalRepoDetail(resp *pb.GetRepoResponse) map[string]any {
	repo := map[string]any{
		"id":             resp.Repo.Id,
		"owner_id":       resp.Repo.OwnerId,
		"name":           resp.Repo.Name,
		"visibility":     resp.Repo.Visibility.String(),
		"description":    resp.Repo.Description,
		"article_count":  resp.Repo.ArticleCount,
		"follower_count": resp.Repo.FollowerCount,
	}
	myRole := resp.MyRole.String()
	if myRole == "COLLABORATOR_ROLE_UNSPECIFIED" {
		myRole = ""
	}
	return map[string]any{
		KeyRepo:        repo,
		"my_role":      myRole,
		"is_following": resp.IsFollowing,
	}
}

// InternalCreateFile 创建文件（内部服务间调用）
// POST /internal/repos/files?owner_id=&repo_id=&path=.
func (h *Handler) InternalCreateFile() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ownerID := ctx.Query("owner_id")
		repoID := ctx.Query("repo_id")
		filePath := ctx.Query("path")
		if ownerID == "" || repoID == "" || filePath == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgOwnerRepoPathReq})
			return
		}
		var req struct {
			Content string `json:"content"`
		}
		if err := ctx.BindAndValidate(&req); err != nil {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgInvalidBody})
			return
		}

		subpath := filepath.Join(ownerID, repoID, filePath)
		if err := h.store.WriteFileFromBytes(subpath, []byte(req.Content)); err != nil {
			slog.Error("创建文件失败", "error", err, "path", subpath)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "创建文件失败"})
			return
		}

		// 异步触发向量化
		if isTextFile(strings.ToLower(filepath.Ext(filePath))) {
			go h.triggerVectorize(c, ownerID, repoID, filePath)
		}

		ctx.JSON(consts.StatusOK, map[string]any{KeyPath: filePath})
	}
}

// InternalUpdateFile 更新文件内容（内部服务间调用）
// PUT /internal/repos/files?owner_id=&repo_id=&path=.
func (h *Handler) InternalUpdateFile() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ownerID := ctx.Query("owner_id")
		repoID := ctx.Query("repo_id")
		filePath := ctx.Query("path")
		if ownerID == "" || repoID == "" || filePath == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgOwnerRepoPathReq})
			return
		}
		var req struct {
			Content string `json:"content"`
		}
		if err := ctx.BindAndValidate(&req); err != nil {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgInvalidBody})
			return
		}

		subpath := filepath.Join(ownerID, repoID, filePath)
		if err := h.store.WriteFileFromBytes(subpath, []byte(req.Content)); err != nil {
			slog.Error("更新文件失败", "error", err, "path", subpath)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "更新文件失败"})
			return
		}

		// 异步触发向量化（StoreChunks 会自动删除旧向量再添加新向量）
		if isTextFile(strings.ToLower(filepath.Ext(filePath))) {
			go h.triggerVectorize(c, ownerID, repoID, filePath)
		}

		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "updated"})
	}
}

// InternalDeleteFile 删除文件（内部服务间调用）
// DELETE /internal/repos/files?owner_id=&repo_id=&path=.
func (h *Handler) InternalDeleteFile() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ownerID := ctx.Query("owner_id")
		repoID := ctx.Query("repo_id")
		filePath := ctx.Query("path")
		if ownerID == "" || repoID == "" || filePath == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgOwnerRepoPathReq})
			return
		}

		subpath := filepath.Join(ownerID, repoID, filePath)
		if err := h.store.Delete(subpath); err != nil {
			slog.Error("删除文件失败", "error", err, "path", subpath)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "删除文件失败"})
			return
		}

		// 异步删除向量
		go h.triggerDeleteFileVectors(c, repoID, filePath)

		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "deleted"})
	}
}

// triggerDeleteFileVectors 异步删除文件向量.
func (h *Handler) triggerDeleteFileVectors(c context.Context, repoID, filePath string) {
	conn := h.grpcClient.GetConn("ai_server")
	if conn == nil {
		slog.Warn("AI 服务连接不可用，跳过删除向量", "file_path", filePath)
		return
	}
	client := pb.NewAIServiceClient(conn)
	_, err := client.DeleteFileVectors(c, &pb.DeleteFileVectorsRequest{
		RepoId:   repoID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Warn("删除向量请求失败", "file_path", filePath, "error", err)
		return
	}

	slog.Info("文件向量已删除", "file_path", filePath)
}

// InternalRenameFile 重命名文件（内部服务间调用）
// PUT /internal/repos/files/rename?owner_id=&repo_id=&path=.
func (h *Handler) InternalRenameFile() app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		ownerID := ctx.Query("owner_id")
		repoID := ctx.Query("repo_id")
		oldPath := ctx.Query("path")
		if ownerID == "" || repoID == "" || oldPath == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgOwnerRepoPathReq})
			return
		}
		var req struct {
			NewPath string `json:"new_path"`
		}
		if err := ctx.BindAndValidate(&req); err != nil || req.NewPath == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: "new_path is required"})
			return
		}

		oldSubpath := filepath.Join(ownerID, repoID, oldPath)
		newSubpath := filepath.Join(ownerID, repoID, req.NewPath)

		if err := h.store.Rename(oldSubpath, newSubpath); err != nil {
			slog.Error("重命名失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "重命名失败"})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			"old_path": oldPath,
			"new_path": req.NewPath,
		})
	}
}

// InternalCreateRepo 创建知识库（内部服务间调用）
// POST /internal/repos.
func (h *Handler) InternalCreateRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Visibility  string `json:"visibility"`
		}
		if err := ctx.BindAndValidate(&req); err != nil || req.Name == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: "name is required"})
			return
		}
		vis := pb.RepoVisibility_PRIVATE
		if req.Visibility == KeyVisibilityPUBLIC {
			vis = pb.RepoVisibility_PUBLIC
		}
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}
		client := pb.NewRepoServiceClient(conn)
		resp, err := client.CreateRepo(c, &pb.CreateRepoRequest{
			OwnerId:     uid,
			Name:        req.Name,
			Description: req.Description,
			Visibility:  vis,
		})
		if err != nil {
			slog.Error("创建知识库失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "创建知识库失败"})
			return
		}
		ctx.JSON(consts.StatusOK, map[string]any{"repo_id": resp.Repo.GetId()})
	}
}

// InternalUpdateRepo 更新知识库（内部服务间调用）
// PUT /internal/repos?repo_id=.
func (h *Handler) InternalUpdateRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgRepoIDRequired})
			return
		}
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Visibility  string `json:"visibility"`
		}
		if err := ctx.BindAndValidate(&req); err != nil {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgInvalidBody})
			return
		}
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}
		client := pb.NewRepoServiceClient(conn)
		updateReq := &pb.UpdateRepoRequest{
			RepoId:      repoID,
			UserId:      uid,
			Name:        req.Name,
			Description: req.Description,
		}
		if req.Visibility == KeyVisibilityPUBLIC {
			updateReq.Visibility = pb.RepoVisibility_PUBLIC
		} else {
			updateReq.Visibility = pb.RepoVisibility_PRIVATE
		}
		_, err := client.UpdateRepo(c, updateReq)
		if err != nil {
			slog.Error("更新知识库失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "更新知识库失败"})
			return
		}
		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "updated"})
	}
}

// InternalSearchUsers 搜索用户（内部服务间调用）
// GET /internal/users/search?q=.
func (h *Handler) InternalSearchUsers() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		keyword := ctx.Query("q")
		if keyword == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: "q is required"})
			return
		}
		conn := h.grpcClient.GetConn("chat_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "chat server unavailable"})
			return
		}
		client := pb.NewChatServiceClient(conn)
		resp, err := client.SearchUsers(c, &pb.SearchUsersReq{Query: keyword})
		if err != nil {
			slog.Error("搜索用户失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "搜索用户失败"})
			return
		}
		users := make([]map[string]any, 0, len(resp.Users))
		for _, u := range resp.Users {
			users = append(users, map[string]any{
				"id":   u.Id,
				"name": u.Name,
			})
		}
		ctx.JSON(consts.StatusOK, map[string]any{"users": users})
	}
}

// InternalAddCollaborator 添加协作者（内部服务间调用）
// POST /internal/repos/collaborators?repo_id=.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) InternalAddCollaborator() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgRepoIDRequired})
			return
		}
		var req struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		}
		if err := ctx.BindAndValidate(&req); err != nil || req.UserID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: "user_id is required"})
			return
		}
		role := pb.CollaboratorRole_VIEWER
		switch req.Role {
		case KeyRoleADMIN:
			role = pb.CollaboratorRole_ADMIN
		case KeyRoleDEVELOPER:
			role = pb.CollaboratorRole_DEVELOPER
		}
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}
		client := pb.NewRepoServiceClient(conn)
		_, err := client.AddCollaborator(c, &pb.AddCollaboratorRequest{
			RepoId:     repoID,
			OperatorId: uid,
			UserId:     req.UserID,
			Role:       role,
		})
		if err != nil {
			slog.Error("添加协作者失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "添加协作者失败"})
			return
		}
		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "added"})
	}
}

// InternalRemoveCollaborator 移除协作者（内部服务间调用）
// DELETE /internal/repos/collaborators?repo_id=&user_id=.
func (h *Handler) InternalRemoveCollaborator() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		targetUserID := ctx.Query("user_id")
		if repoID == "" || targetUserID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: "repo_id and user_id are required"})
			return
		}
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}
		client := pb.NewRepoServiceClient(conn)
		_, err := client.RemoveCollaborator(c, &pb.RemoveCollaboratorRequest{
			RepoId:     repoID,
			OperatorId: uid,
			UserId:     targetUserID,
		})
		if err != nil {
			slog.Error("移除协作者失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "移除协作者失败"})
			return
		}
		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "removed"})
	}
}

// InternalUpdateCollaboratorRole 更新协作者角色（内部服务间调用）
// PUT /internal/repos/collaborators/role?repo_id=.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) InternalUpdateCollaboratorRole() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgRepoIDRequired})
			return
		}
		var req struct {
			UserID string `json:"user_id"`
			Role   string `json:"role"`
		}
		if err := ctx.BindAndValidate(&req); err != nil || req.UserID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: "user_id is required"})
			return
		}
		role := pb.CollaboratorRole_VIEWER
		switch req.Role {
		case KeyRoleADMIN:
			role = pb.CollaboratorRole_ADMIN
		case KeyRoleDEVELOPER:
			role = pb.CollaboratorRole_DEVELOPER
		}
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}
		client := pb.NewRepoServiceClient(conn)
		_, err := client.UpdateCollaborator(c, &pb.UpdateCollaboratorRequest{
			RepoId:     repoID,
			OperatorId: uid,
			UserId:     req.UserID,
			Role:       role,
		})
		if err != nil {
			slog.Error("更新协作者角色失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "更新协作者角色失败"})
			return
		}
		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "updated"})
	}
}

// InternalListFollowedRepos 获取用户关注的知识库列表（内部服务间调用）
// GET /internal/repos/followed.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) InternalListFollowedRepos() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		if uid == "" {
			ctx.JSON(consts.StatusUnauthorized, map[string]string{KeyMessage: ErrMsgMissingUserID})
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListFollowedRepos(c, &pb.ListFollowedReposRequest{UserId: uid})
		if err != nil {
			slog.Error("获取关注列表失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "获取关注列表失败"})
			return
		}

		repos := make([]map[string]any, 0, len(resp.Repos))
		for _, r := range resp.Repos {
			repos = append(repos, marshalInternalRepo(r))
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			KeyRepos: repos,
			"total":  len(repos),
		})
	}
}

// InternalFollowRepo 关注知识库（内部服务间调用）
// POST /internal/repos/follow?repo_id=xxx.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) InternalFollowRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgRepoIDRequired})
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		_, err := client.FollowRepo(c, &pb.FollowRepoRequest{RepoId: repoID, UserId: uid})
		if err != nil {
			slog.Error("关注知识库失败", "error", err, "repo_id", repoID)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "关注知识库失败"})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "followed"})
	}
}

// InternalUnfollowRepo 取消关注知识库（内部服务间调用）
// DELETE /internal/repos/follow?repo_id=xxx.
//
//nolint:dupl // handler 结构一致是 gateway 层自然模式
func (h *Handler) InternalUnfollowRepo() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgRepoIDRequired})
			return
		}

		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}

		client := pb.NewRepoServiceClient(conn)
		_, err := client.UnfollowRepo(c, &pb.UnfollowRepoRequest{RepoId: repoID, UserId: uid})
		if err != nil {
			slog.Error("取消关注知识库失败", "error", err, "repo_id", repoID)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "取消关注知识库失败"})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]any{KeyMessage: "unfollowed"})
	}
}

// InternalListCollaborators 列出协作者（内部服务间调用）
// GET /internal/repos/collaborators?repo_id=.
func (h *Handler) InternalListCollaborators() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Query("repo_id")
		if repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{KeyMessage: ErrMsgRepoIDRequired})
			return
		}
		conn := h.grpcClient.GetConn("repo_server")
		if conn == nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: ErrMsgRepoServerUnavail})
			return
		}
		client := pb.NewRepoServiceClient(conn)
		resp, err := client.ListCollaborators(c, &pb.ListCollaboratorsRequest{
			RepoId: repoID,
			UserId: uid,
		})
		if err != nil {
			slog.Error("列出协作者失败", "error", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]string{KeyMessage: "列出协作者失败"})
			return
		}
		items := make([]map[string]any, 0, len(resp.Collaborators))
		for _, c := range resp.Collaborators {
			items = append(items, map[string]any{
				KeyUserID: c.UserId,
				KeyRole:   c.Role.String(),
			})
		}
		ctx.JSON(consts.StatusOK, map[string]any{
			"collaborators": items,
		})
	}
}
