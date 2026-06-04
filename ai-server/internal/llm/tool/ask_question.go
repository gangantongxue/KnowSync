package tool

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// OnAskUser 反问用户回调，当工具被调用时触发，传入工具返回的 JSON 结果.
type OnAskUser func(resultJSON string)

// AskQuestion 反问用户工具，实现 Eino InvokableTool 接口
// 当 LLM 需要更多信息才能回答用户问题时，通过此工具向用户提问.
type AskQuestion struct {
	onAskUser OnAskUser
}

// NewAskQuestion 创建反问用户工具
// onAskUser: 可选，当非 nil 时，工具被调用后会回调此函数.
func NewAskQuestion(onAskUser OnAskUser) *AskQuestion {
	return &AskQuestion{onAskUser: onAskUser}
}

// Info 返回工具元信息.
func (a *AskQuestion) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "ask_user",
		Desc: "当你需要更多信息才能回答用户问题时，使用此工具向用户提问。支持单选（single）和多选（multiple）。options 中应提供完整选项供用户选择，系统会自动在选项中追加自由输入选项。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"question": {
				Type:     TypeString,
				Desc:     "向用户提出的问题，描述需要什么信息",
				Required: true,
			},
			ParamType: {
				Type:     TypeString,
				Desc:     "选项类型：single-单选, multiple-多选",
				Required: true,
			},
			"options": {
				Type:     "array",
				Desc:     "提供的选项列表，每个选项为一个字符串",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行工具调用.
func (a *AskQuestion) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	result, err := executeAskQuestion(ctx, arguments)
	if err == nil && a.onAskUser != nil {
		a.onAskUser(result)
	}
	return result, err
}

// AskQuestionParams 工具参数.
type AskQuestionParams struct {
	Question string   `json:"question"`
	Type     string   `json:"type"` // "single" 或 "multiple"
	Options  []string `json:"options"`
}

func executeAskQuestion(ctx context.Context, paramsJSON string) (string, error) {
	_ = ctx

	var params AskQuestionParams
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", err
	}

	if params.Question == "" {
		return `{"action": "ask_user_error", "message": "问题不能为空"}`, nil
	}
	if params.Type != "single" && params.Type != "multiple" {
		params.Type = "single"
	}

	data, _ := json.Marshal(map[string]any{
		"action":   "ask_user",
		"question": params.Question,
		ParamType:  params.Type,
		"options":  params.Options,
	})
	return string(data), nil
}
