package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// Logging 请求日志中间件，记录 method、path、status、耗时
func Logging() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()
		path := string(ctx.Request.URI().Path())
		method := string(ctx.Request.Method())

		ctx.Next(c)

		latency := time.Since(start)
		status := ctx.Response.StatusCode()

		slog.Info("request",
			"method", method,
			"path", path,
			"status", status,
			"latency", latency.String(),
		)
	}
}
