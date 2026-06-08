package handler

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
)

// RenameFileReq 文件重命名/移动请求体.
type RenameFileReq struct {
	NewPath string `json:"new_path"`
}

// GetRepoTree 获取知识库文件树
// GET /api/v1/repos/:repo_id/tree?path=docs/.
func (h *Handler) GetRepoTree() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		dirPath := ctx.Query("path")

		subpath := filepath.Join(uid, repoID, dirPath)
		entries, err := h.store.ListDir(subpath)
		if err != nil {
			response.Error(c, ctx, 404, errcode.ErrNotFound, "目录不存在")
			return
		}

		type treeEntry struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Size int64  `json:"size"`
		}
		result := make([]treeEntry, 0, len(entries))
		for _, e := range entries {
			t := "file"
			if e.IsDir {
				t = "dir"
			}
			result = append(result, treeEntry{Name: e.Name, Type: t, Size: e.Size})
		}

		response.Success(c, ctx, map[string]any{
			"entries": result,
			KeyPath:   dirPath,
		})
	}
}

// UploadFile 上传文件到知识库
// POST /api/v1/repos/:repo_id/files?path=docs/chapter1.md.
func (h *Handler) UploadFile() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		filePath := ctx.Query("path")
		if filePath == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "path 参数必填")
			return
		}

		fileHeader, err := ctx.FormFile("file")
		if err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "文件上传失败")
			return
		}

		f, err := fileHeader.Open()
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "文件读取失败")
			return
		}
		defer func() { _ = f.Close() }()

		subpath := filepath.Join(uid, repoID, filePath)
		size, err := h.store.WriteFile(subpath, f)
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "文件存储失败")
			return
		}

		// 异步触发向量化（仅对文本文件）
		ext := strings.ToLower(filepath.Ext(filePath))
		if isTextFile(ext) {
			go h.triggerVectorize(c, uid, repoID, filePath)
		}

		response.Success(c, ctx, map[string]any{
			"path":     filePath,
			"size":     size,
			"file_url": fmt.Sprintf("files/%s/%s/%s", uid, repoID, filePath),
		})
	}
}

// triggerVectorize 异步触发向量化.
func (h *Handler) triggerVectorize(c context.Context, uid, repoID, filePath string) {
	conn := h.grpcClient.GetConn("ai_server")
	if conn == nil {
		slog.Warn("AI 服务连接不可用，跳过向量化", "file_path", filePath)
		return
	}

	client := pb.NewAIServiceClient(conn)
	_, err := client.VectorizeArticle(c, &pb.VectorizeArticleRequest{
		UserId:   uid,
		RepoId:   repoID,
		FilePath: filePath,
	})
	if err != nil {
		slog.Warn("向量化请求失败", "file_path", filePath, "error", err)
		return
	}

	slog.Info("向量化任务已提交", "file_path", filePath)
}

// isTextFile 判断是否为需要向量化的文本文件.
func isTextFile(ext string) bool {
	switch ext {
	case ".md", ".txt", ".markdown", ".rst", ".adoc", ".asciidoc":
		return true
	}
	return false
}

// DeleteFile 删除知识库中的文件或目录
// DELETE /api/v1/repos/:repo_id/files?path=docs/old.md.
func (h *Handler) DeleteFile() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		filePath := ctx.Query("path")
		if filePath == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "path 参数必填")
			return
		}

		subpath := filepath.Join(uid, repoID, filePath)

		info, err := h.store.Stat(subpath)
		if err != nil {
			response.Error(c, ctx, 404, errcode.ErrNotFound, "文件不存在")
			return
		}

		if info.IsDir() {
			if err := h.store.DeleteAll(subpath); err != nil {
				response.Error(c, ctx, 500, errcode.ErrBadReq, "删除目录失败")
				return
			}
		} else {
			if err := h.store.Delete(subpath); err != nil {
				response.Error(c, ctx, 500, errcode.ErrBadReq, "删除文件失败")
				return
			}
		}

		response.Success(c, ctx, nil)
	}
}

// RenameFile 重命名或移动文件/目录
// PUT /api/v1/repos/:repo_id/files?path=docs/old.md  body: { "new_path": "docs/new.md" }.
func (h *Handler) RenameFile() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		oldPath := ctx.Query("path")
		if oldPath == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "path 参数必填")
			return
		}

		var req RenameFileReq
		if err := ctx.BindAndValidate(&req); err != nil {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "请求参数错误")
			return
		}
		if req.NewPath == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "new_path 必填")
			return
		}

		oldSubpath := filepath.Join(uid, repoID, oldPath)
		newSubpath := filepath.Join(uid, repoID, req.NewPath)

		if err := h.store.Rename(oldSubpath, newSubpath); err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "重命名失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			"old_path": oldPath,
			"new_path": req.NewPath,
		})
	}
}

// MakeDir 在知识库中创建目录
// POST /api/v1/repos/:repo_id/dirs?path=docs/new-folder.
func (h *Handler) MakeDir() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid := ctx.GetString("user_id")
		repoID := ctx.Param("repo_id")
		dirPath := ctx.Query("path")
		if dirPath == "" {
			response.Error(c, ctx, 400, errcode.ErrBadReq, "path 参数必填")
			return
		}

		subpath := filepath.Join(uid, repoID, dirPath)
		if err := h.store.MakeDir(subpath); err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "创建目录失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			"path": dirPath,
		})
	}
}

// InternalRepoTree 内部服务间获取仓库文件树（无鉴权，用于 ai-server 的 LLM 工具）
// GET /internal/repos/tree?owner_id=&repo_id=&path=.
func InternalRepoTree(store *storage.Store) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		ownerID := ctx.Query("owner_id")
		repoID := ctx.Query("repo_id")
		dirPath := ctx.Query("path")

		if ownerID == "" || repoID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{
				KeyCode:    "MISSING_PARAM",
				KeyMessage: "owner_id and repo_id are required",
			})
			return
		}

		subpath := filepath.Join(ownerID, repoID, dirPath)
		entries, err := store.ListDir(subpath)
		if err != nil {
			ctx.JSON(consts.StatusNotFound, map[string]string{
				"code":     "DIR_NOT_FOUND",
				KeyMessage: "directory not found",
			})
			return
		}

		type treeEntry struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Size int64  `json:"size"`
		}
		result := make([]treeEntry, 0, len(entries))
		for _, e := range entries {
			t := "file"
			if e.IsDir {
				t = "dir"
			}
			result = append(result, treeEntry{Name: e.Name, Type: t, Size: e.Size})
		}

		ctx.JSON(consts.StatusOK, map[string]any{
			"entries": result,
			KeyPath:   dirPath,
		})
	}
}
