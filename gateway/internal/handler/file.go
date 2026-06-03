package handler

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
)

// RenameFileReq 文件重命名/移动请求体
type RenameFileReq struct {
	NewPath string `json:"new_path"`
}

// GetRepoTree 获取知识库文件树
// GET /api/v1/repos/:repo_id/tree?path=docs/
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
			Name  string `json:"name"`
			Type  string `json:"type"`
			Size  int64  `json:"size"`
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
			"path":    dirPath,
		})
	}
}

// UploadFile 上传文件到知识库
// POST /api/v1/repos/:repo_id/files?path=docs/chapter1.md
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
		defer f.Close()

		subpath := filepath.Join(uid, repoID, filePath)
		size, err := h.store.WriteFile(subpath, f)
		if err != nil {
			response.Error(c, ctx, 500, errcode.ErrBadReq, "文件存储失败")
			return
		}

		response.Success(c, ctx, map[string]any{
			"path":     filePath,
			"size":     size,
			"file_url": fmt.Sprintf("files/%s/%s/%s", uid, repoID, filePath),
		})
	}
}

// DeleteFile 删除知识库中的文件或目录
// DELETE /api/v1/repos/:repo_id/files?path=docs/old.md
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
// PUT /api/v1/repos/:repo_id/files?path=docs/old.md  body: { "new_path": "docs/new.md" }
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
// POST /api/v1/repos/:repo_id/dirs?path=docs/new-folder
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
