package handler

import (
	"context"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gangantongxue/knowsync/gateway/pkg/storage"
)

// FileHandler 创建文件下载请求的 HTTP 处理器
// 路由:
//
//	GET /files/public/*filepath — 公共读，直接返回文件
//	GET /files/auth/:token     — 认证读，解析 JWT 后返回文件
//
// 底层使用 ctx.File()，原生支持 Range 请求头 (206 Partial Content/断点续传)
func FileHandler(store *storage.Store) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		bucket := storage.Bucket(ctx.Param("bucket"))

		switch bucket {
		case storage.BucketPublic:
			serveFile(ctx, store, bucket, ctx.Param("filepath"))

		case storage.BucketAuth:
			claims, err := store.ValidateToken(ctx.Param("token"))
			if err != nil {
				ctx.JSON(consts.StatusUnauthorized, map[string]string{
					"code":    "INVALID_TOKEN",
					"message": err.Error(),
				})
				return
			}
			serveFile(ctx, store, storage.Bucket(claims.Bucket), claims.Path)

		default:
			ctx.JSON(consts.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "bucket not found",
			})
		}
	}
}

// serveFile 校验路径并返回文件内容
func serveFile(ctx *app.RequestContext, store *storage.Store, bucket storage.Bucket, key string) {
	fullPath, err := store.ResolvePath(bucket, key)
	if err != nil {
		ctx.JSON(consts.StatusBadRequest, map[string]string{
			"code":    "INVALID_PATH",
			"message": err.Error(),
		})
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		ctx.JSON(consts.StatusNotFound, map[string]string{
			"code":    "FILE_NOT_FOUND",
			"message": "file not found",
		})
		return
	}

	ctx.File(fullPath)
}
