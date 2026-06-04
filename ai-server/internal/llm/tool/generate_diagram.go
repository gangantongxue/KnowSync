package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

//nolint:revive // self-documenting
type GenerateDiagram struct {
	fileWriteClient FileWriteClient
}

//nolint:revive // self-documenting
func NewGenerateDiagram(fwc FileWriteClient) *GenerateDiagram {
	return &GenerateDiagram{fileWriteClient: fwc}
}

var validDiagramTypes = map[string]string{
	"flowchart": "流程图（flowchart）",
	"sequence":  "时序图（sequenceDiagram）",
	"class":     "类图（classDiagram）",
	"state":     "状态图（stateDiagram）",
	"er":        "ER 图（erDiagram）",
	"gantt":     "甘特图（gantt）",
	"pie":       "饼图（pie）",
	"mindmap":   "思维导图（mindmap）",
	"timeline":  "时间线（timeline）",
}

//nolint:revive // self-documenting
func (g *GenerateDiagram) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "generate_diagram",
		Desc: "生成 Mermaid 图表并保存为知识库文件。支持的图表类型：flowchart（流程图）、sequence（时序图）、class（类图）、state（状态图）、er（ER 图）、gantt（甘特图）、pie（饼图）、mindmap（思维导图）、timeline（时间线）。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			ParamType: {
				Type:     TypeString,
				Desc:     "图表类型：" + joinDiagramTypes(),
				Required: true,
			},
			"code": {
				Type:     TypeString,
				Desc:     "Mermaid 图表代码（合法的 Mermaid 语法）",
				Required: true,
			},
			ParamRepoID: {
				Type:     TypeString,
				Desc:     "保存到的知识库 ID",
				Required: true,
			},
			ParamFilePath: {
				Type:     TypeString,
				Desc:     "保存的文件路径（可选，默认自动生成）",
				Required: false,
			},
			ParamSkipCfm: {
				Type:     TypeBoolean,
				Desc:     DescSkipConfirm,
				Required: false,
			},
		}),
	}, nil
}

//nolint:revive // self-documenting
func (g *GenerateDiagram) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return g.execute(ctx, arguments)
}

func (g *GenerateDiagram) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		Type     string `json:"type"`
		Code     string `json:"code"`
		RepoID   string `json:"repo_id"`
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.Type == "" {
		return `{"success": false, "message": "type 不能为空"}`, nil
	}
	if params.Code == "" {
		return `{"success": false, "message": "code 不能为空"}`, nil
	}
	if params.RepoID == "" {
		return ErrRespRepoIDEmpty, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return ErrRespUserInfo, nil
	}

	// 校验图表类型
	diagramType, valid := validDiagramTypes[params.Type]
	if !valid {
		types := make([]string, 0, len(validDiagramTypes))
		for k := range validDiagramTypes {
			types = append(types, k)
		}
		return fmt.Sprintf(`{"success": false, "message": "不支持的图表类型: %s，支持的类型: %s"}`, params.Type, strings.Join(types, ", ")), nil
	}

	// 自动生成文件路径
	filePath := params.FilePath
	if filePath == "" {
		filePath = fmt.Sprintf("diagrams/%s_%s.mmd", params.Type, generateShortName(params.Code))
	}

	slog.Info("生成图表", "type", params.Type, "repo_id", params.RepoID, "file_path", filePath)

	// 构建完整的 Mermaid 文件内容
	mermaidHeader := mermaidHeaderForType(params.Type)
	content := mermaidHeader + "\n" + params.Code

	// 基本语法校验
	if err := validateMermaid(params.Type, params.Code); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "Mermaid 语法校验失败: %s", "code": %s}`, err.Error(), jsonString(params.Code)), nil
	}

	// 保存文件
	if err := g.fileWriteClient.CreateFile(ctx, userID, params.RepoID, filePath, content); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "保存文件失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		KeySuccess:    true,
		ParamType:     params.Type,
		"type_name":   diagramType,
		ParamRepoID:   params.RepoID,
		ParamFilePath: filePath,
		"code_length": len(params.Code),
	})
	return string(data), nil
}

func mermaidHeaderForType(t string) string {
	switch t {
	case "flowchart":
		return "```mermaid\nflowchart TD"
	case "sequence":
		return "```mermaid\nsequenceDiagram"
	case "class":
		return "```mermaid\nclassDiagram"
	case "state":
		return "```mermaid\nstateDiagram-v2"
	case "er":
		return "```mermaid\nerDiagram"
	case "gantt":
		return "```mermaid\ngantt"
	case "pie":
		return "```mermaid\npie"
	case "mindmap":
		return "```mermaid\nmindmap"
	case "timeline":
		return "```mermaid\ntimeline"
	default:
		return "```mermaid"
	}
}

func validateMermaid(diagramType, code string) error {
	_ = diagramType
	lines := strings.Split(code, "\n")
	nonEmpty := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "%%") {
			nonEmpty++
		}
	}
	if nonEmpty == 0 {
		return errors.New("图表内容为空")
	}

	// 检查括号匹配
	depth := 0
	for _, r := range code {
		switch r {
		case '(', '{', '[':
			depth++
		case ')', '}', ']':
			depth--
			if depth < 0 {
				return errors.New("括号不匹配：存在多余的右括号")
			}
		}
	}
	if depth > 0 {
		return errors.New("括号不匹配：存在未闭合的左括号")
	}

	return nil
}

//nolint:gocyclo // 需要处理多种图表类型和格式
func generateShortName(code string) string {
	// 取第一行非空内容的前 20 个字符作为文件名
	for line := range strings.SplitSeq(code, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "%%") && !strings.HasPrefix(trimmed, "---") {
			name := strings.TrimSpace(trimmed)
			name = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
					return r
				}
				return '_'
			}, name)
			if len(name) > 30 {
				name = name[:30]
			}
			if name == "" {
				name = "diagram"
			}
			return strings.ToLower(name)
		}
	}
	return "diagram"
}

func joinDiagramTypes() string {
	types := make([]string, 0, len(validDiagramTypes))
	for k, v := range validDiagramTypes {
		types = append(types, fmt.Sprintf("%s（%s）", k, v))
	}
	return strings.Join(types, "、")
}

func jsonString(s string) string {
	data, _ := json.Marshal(s)
	return string(data)
}
