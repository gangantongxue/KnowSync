package response

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Response 统一 HTTP JSON 响应格式
// code 为业务状态码，200 表示成功，非 200 表示各类错误
// message 为提示信息，data 为响应数据（失败时为 null）
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Success 返回成功响应（HTTP 200）
func Success(c context.Context, ctx *app.RequestContext, data any) {
	ctx.JSON(consts.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应，httpStatus 为 HTTP 状态码，code 为业务错误码
func Error(c context.Context, ctx *app.RequestContext, httpStatus, code int, message string) {
	ctx.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
