package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// Calculator 数学运算工具.
type Calculator struct{}

// NewCalculator 创建 Calculator 工具.
func NewCalculator() *Calculator {
	return &Calculator{}
}

//nolint:revive // self-documenting
//nolint:revive // self-documenting
func (c *Calculator) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行精确的数学运算。支持四则运算（+ - * /）、幂运算（**）、取模（%%）、三角函数（sin/cos/tan）、对数（log/ln）、平方根（sqrt）、绝对值（abs）、取整（ceil/floor/round）。常量支持 pi、e。例如：sqrt(sin(pi/4)^2 + cos(pi/4)^2)",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     TypeString,
				Desc:     "要计算的数学表达式",
				Required: true,
			},
		}),
	}, nil
}

//nolint:revive // self-documenting
func (c *Calculator) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return c.execute(ctx, arguments)
}

func (c *Calculator) execute(ctx context.Context, paramsJSON string) (string, error) {
	_ = ctx

	var params struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.Expression == "" {
		return `{"success": false, "error": "表达式不能为空"}`, nil
	}

	result, err := eval(params.Expression)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "error": "计算错误: %s"}`, err.Error()), nil
	}

	return fmt.Sprintf(`{"success": true, "expression": "%s", "result": %s}`,
		params.Expression, formatNumber(result)), nil
}

func formatNumber(v float64) string {
	if v == math.Trunc(v) && !math.IsInf(v, 0) && math.Abs(v) < 1e15 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.10g", v)
}

func eval(expr string) (float64, error) {
	exprObj, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, fmt.Errorf("无法解析表达式: %w", err)
	}
	return evalExpr(exprObj)
}

//nolint:gocyclo // 表达式解析需要处理多种 ast 节点类型
func evalExpr(n ast.Expr) (float64, error) {
	switch e := n.(type) {
	case *ast.BasicLit:
		return strconv.ParseFloat(e.Value, 64)

	case *ast.ParenExpr:
		return evalExpr(e.X)

	case *ast.UnaryExpr:
		x, err := evalExpr(e.X)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case token.SUB:
			return -x, nil
		case token.ADD:
			return x, nil
		default:
			return 0, fmt.Errorf("不支持的一元运算符: %s", e.Op)
		}

	case *ast.BinaryExpr:
		x, err := evalExpr(e.X)
		if err != nil {
			return 0, err
		}
		y, err := evalExpr(e.Y)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case token.ADD:
			return x + y, nil
		case token.SUB:
			return x - y, nil
		case token.MUL:
			return x * y, nil
		case token.QUO:
			if y == 0 {
				return 0, errors.New("除数不能为 0")
			}
			return x / y, nil
		case token.REM:
			return float64(int64(x) % int64(y)), nil
		default:
			return 0, fmt.Errorf("不支持的二元运算符: %s", e.Op)
		}

	case *ast.CallExpr:
		return evalCall(e)

	case *ast.Ident:
		switch e.Name {
		case "pi":
			return math.Pi, nil
		case "e":
			return math.E, nil
		}
		return 0, fmt.Errorf("未知常量: %s", e.Name)

	default:
		return 0, fmt.Errorf("不支持的表达式类型: %T", n)
	}
}

//nolint:gocyclo // 函数调用分发需要处理多种数学函数
func evalCall(call *ast.CallExpr) (float64, error) {
	fn, ok := call.Fun.(*ast.Ident)
	if !ok {
		return 0, errors.New("不支持的函数调用")
	}

	args := make([]float64, len(call.Args))
	for i, arg := range call.Args {
		v, err := evalExpr(arg)
		if err != nil {
			return 0, err
		}
		args[i] = v
	}

	switch fn.Name {
	case "sqrt":
		if len(args) != 1 {
			return 0, errors.New("sqrt 需要 1 个参数")
		}
		return math.Sqrt(args[0]), nil

	case "sin":
		if len(args) != 1 {
			return 0, errors.New("sin 需要 1 个参数")
		}
		return math.Sin(args[0]), nil

	case "cos":
		if len(args) != 1 {
			return 0, errors.New("cos 需要 1 个参数")
		}
		return math.Cos(args[0]), nil

	case "tan":
		if len(args) != 1 {
			return 0, errors.New("tan 需要 1 个参数")
		}
		return math.Tan(args[0]), nil

	case "asin":
		if len(args) != 1 {
			return 0, errors.New("asin 需要 1 个参数")
		}
		return math.Asin(args[0]), nil

	case "acos":
		if len(args) != 1 {
			return 0, errors.New("acos 需要 1 个参数")
		}
		return math.Acos(args[0]), nil

	case "atan":
		if len(args) != 1 {
			return 0, errors.New("atan 需要 1 个参数")
		}
		return math.Atan(args[0]), nil

	case "log":
		if len(args) != 1 {
			return 0, errors.New("log 需要 1 个参数")
		}
		return math.Log10(args[0]), nil

	case "ln":
		if len(args) != 1 {
			return 0, errors.New("ln 需要 1 个参数")
		}
		return math.Log(args[0]), nil

	case "abs":
		if len(args) != 1 {
			return 0, errors.New("abs 需要 1 个参数")
		}
		return math.Abs(args[0]), nil

	case "ceil":
		if len(args) != 1 {
			return 0, errors.New("ceil 需要 1 个参数")
		}
		return math.Ceil(args[0]), nil

	case "floor":
		if len(args) != 1 {
			return 0, errors.New("floor 需要 1 个参数")
		}
		return math.Floor(args[0]), nil

	case "round":
		if len(args) != 1 {
			return 0, errors.New("round 需要 1 个参数")
		}
		return math.Round(args[0]), nil

	case "pow":
		if len(args) != 2 {
			return 0, errors.New("pow 需要 2 个参数")
		}
		return math.Pow(args[0], args[1]), nil

	default:
		return 0, fmt.Errorf("不支持的函数: %s", fn.Name)
	}
}
