package tool

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/compose"
)

// CtxKeyConfirmed 用于在 context 中传递确认标记.
const CtxKeyConfirmed = "confirmed_write"

// writeToolNames 所有需要确认的写操作工具名.
var writeToolNames = map[string]bool{
	"create_file":              true,
	"update_file":              true,
	"delete_file":              true,
	"rename_file":              true,
	"create_repo":              true,
	"update_repo":              true,
	"add_collaborator":         true,
	"remove_collaborator":      true,
	"update_collaborator_role": true,
}

// NewConfirmationMiddleware 创建写操作确认中间件
//
// policies: 各个工具的确认策略
// 对于 ConfirmAlways — 始终拦截，强制要求确认
// 对于 ConfirmOptional — 检查参数中是否有 _skip_confirm: true，有则跳过.
func NewConfirmationMiddleware(policies map[string]ConfirmLevel) compose.InvokableToolMiddleware {
	return func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
		return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
			name := input.Name

			// 非写操作工具直接放行
			if !writeToolNames[name] {
				return next(ctx, input)
			}

			policy, ok := policies[name]
			if !ok || policy == ConfirmNever {
				return next(ctx, input)
			}

			// 已确认 → 放行
			if ctx.Value(CtxKeyConfirmed) != nil {
				return next(ctx, input)
			}

			var raw map[string]any
			_ = json.Unmarshal([]byte(input.Arguments), &raw)

			// update_repo 的 visibility 变更始终强制确认
			effectivePolicy := policy
			if name == "update_repo" {
				if vis, ok := raw["visibility"].(string); ok && vis != "" {
					effectivePolicy = ConfirmAlways
				}
			}

			// 按需确认 + 参数带跳过标记 → 放行
			if effectivePolicy == ConfirmOptional {
				if skip, _ := raw["_skip_confirm"].(bool); skip {
					return next(ctx, input)
				}
			}

			// 未确认 → 返回确认事件
			result, _ := json.Marshal(map[string]any{
				"action": "confirm_write",
				"tool":   name,
				"params": input.Arguments,
			})
			return &compose.ToolOutput{Result: string(result)}, nil
		}
	}
}
