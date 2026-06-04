package handler

import (
	"context"
	"os"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// FileHandler 公共文件访问处理器
// GET /files/*filepath — 直接返回文件，支持 Range 请求.
func FileHandler(store *storage.Store) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		filepath := ctx.Param("filepath")
		filepath = strings.TrimPrefix(filepath, "/")
		serveFile(ctx, store, filepath)
	}
}

// InternalFileHandler 内部服务间文件读取
// GET /internal/file?path=...
func InternalFileHandler(store *storage.Store) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		path := ctx.Query("path")
		if path == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{
				KeyCode:    "MISSING_PATH",
				KeyMessage: "path is required",
			})
			return
		}
		serveFile(ctx, store, path)
	}
}

func serveFile(ctx *app.RequestContext, store *storage.Store, path string) {
	fullPath, err := store.Resolve(path)
	if err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]string{
			"code":     "INVALID_PATH",
			KeyMessage: err.Error(),
		})
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		ctx.JSON(consts.StatusNotFound, map[string]string{
			"code":     "FILE_NOT_FOUND",
			KeyMessage: "file not found",
		})
		return
	}

	ctx.File(fullPath)
}
