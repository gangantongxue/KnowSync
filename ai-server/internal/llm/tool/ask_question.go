package tool

import (
	"context"
	"encoding/json"
)

// AskQuestion 反问用户工具
// 当 LLM 需要更多信息才能回答用户问题时，通过此工具向用户提问
type AskQuestion struct {
	Tool
}

// NewAskQuestion 创建反问用户工具
func NewAskQuestion() *AskQuestion {
	return &AskQuestion{
		Tool: Tool{
			Name:        "ask_user",
			Description: "当你需要更多信息才能回答用户问题时，使用此工具向用户提问。支持单选和多选，始终包含自由输入选项。请尽量提供完整的选项供用户选择。",
			ParamsJSON:  askQuestionSchema(),
			Handler:     nil,
		},
	}
}

// Init 初始化 Handler
func (a *AskQuestion) Init() *Tool {
	a.Tool.Handler = func(ctx context.Context, paramsJSON string) (string, error) {
		return executeAskQuestion(ctx, paramsJSON)
	}
	return &a.Tool
}

// AskQuestionParams 工具参数
type AskQuestionParams struct {
	Question string   `json:"question"`
	Type     string   `json:"type"` // "single" 或 "multiple"
	Options  []string `json:"options"`
}

// AskQuestionResult 返回给服务层的结果（不会被 LLM 直接使用）
type AskQuestionResult struct {
	Action   string   `json:"action"`
	Question string   `json:"question"`
	Type     string   `json:"type"`
	Options  []string `json:"options"`
}

func executeAskQuestion(ctx context.Context, paramsJSON string) (string, error) {
	var params AskQuestionParams
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return `{"action": "ask_user_error", "message": "参数解析失败"}`, nil
	}

	if params.Question == "" {
		return `{"action": "ask_user_error", "message": "问题不能为空"}`, nil
	}
	if params.Type != "single" && params.Type != "multiple" {
		params.Type = "single"
	}

	result := AskQuestionResult{
		Action:   "ask_user",
		Question: params.Question,
		Type:     params.Type,
		Options:  params.Options,
	}

	data, _ := json.Marshal(result)
	return string(data), nil
}

func askQuestionSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"question": {
				"type": "string",
				"description": "向用户提出的问题，描述需要什么信息"
			},
			"type": {
				"type": "string",
				"enum": ["single", "multiple"],
				"description": "选项类型：single-单选, multiple-多选"
			},
			"options": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "提供的选项列表，每个选项为一个字符串"
			}
		},
		"required": ["question", "type", "options"]
	}`)
}
