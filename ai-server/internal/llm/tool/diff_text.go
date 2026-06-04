package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type edit struct {
	kind string // "eq", "ins", "del"
	line string
}

type DiffText struct{}

func NewDiffText() *DiffText {
	return &DiffText{}
}

func (d *DiffText) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "diff_text",
		Desc: "比较两段文本的差异，返回统一格式（unified diff）的差异对比结果。适用于比较文章的不同版本、代码变更等场景。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text_a": {
				Type:     "string",
				Desc:     "原始文本（旧版本）",
				Required: true,
			},
			"text_b": {
				Type:     "string",
				Desc:     "新文本（新版本）",
				Required: true,
			},
			"context_lines": {
				Type:     "integer",
				Desc:     "上下文行数（可选，默认 3）",
				Required: false,
			},
		}),
	}, nil
}

func (d *DiffText) InvokableRun(ctx context.Context, arguments string, opts ...tool.Option) (string, error) {
	return d.execute(ctx, arguments)
}

func (d *DiffText) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		TextA        string `json:"text_a"`
		TextB        string `json:"text_b"`
		ContextLines int    `json:"context_lines"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.ContextLines <= 0 {
		params.ContextLines = 3
	}

	linesA := strings.Split(params.TextA, "\n")
	linesB := strings.Split(params.TextB, "\n")

	edits := computeDiff(linesA, linesB)

	// 生成 unified diff 格式
	stats := struct {
		Additions int `json:"additions"`
		Deletions int `json:"deletions"`
		Hunks     int `json:"hunks"`
	}{}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("--- original\n+++ modified\n"))

	var hunk []edit
	hunkStartA, hunkStartB := -1, -1

	flushHunk := func() {
		if len(hunk) == 0 {
			return
		}

		stats.Hunks++
		// 计算 hunk 的起始行和范围
		startA := hunkStartA
		startB := hunkStartB
		countA := 0
		countB := 0
		for _, e := range hunk {
			switch e.kind {
			case "eq":
				countA++
				countB++
			case "del":
				countA++
			case "ins":
				countB++
			}
		}

		result.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", startA, countA, startB, countB))
		for _, e := range hunk {
			switch e.kind {
			case "eq":
				result.WriteString(" " + e.line + "\n")
			case "del":
				result.WriteString("-" + e.line + "\n")
				stats.Deletions++
			case "ins":
				result.WriteString("+" + e.line + "\n")
				stats.Additions++
			}
		}
		hunk = nil
	}

	for i := 0; i < len(edits); {
		if edits[i].kind == "eq" {
			// 收集上下文行
			ctxBefore := max(0, i-params.ContextLines)
			ctxAfter := min(len(edits), i+params.ContextLines+1)

			// 检查周围是否有变更
			hasChange := false
			for j := ctxBefore; j < ctxAfter; j++ {
				if edits[j].kind != "eq" {
					hasChange = true
					break
				}
			}

			if hasChange {
				if hunkStartA < 0 {
					// 新 hunk 开始
					hunkStartA = max(1, i+1-params.ContextLines)
					hunkStartB = hunkStartA
					// 添加上下文
					ctxStart := max(0, i-params.ContextLines)
					for j := ctxStart; j < i; j++ {
						hunk = append(hunk, edit{kind: "eq", line: edits[j].line})
					}
				}
				hunk = append(hunk, edits[i])
			} else {
				flushHunk()
				hunkStartA = -1
				hunkStartB = -1
			}
			i++
		} else {
			if hunkStartA < 0 {
				hunkStartA = max(1, i+1-params.ContextLines)
				hunkStartB = hunkStartA
				ctxStart := max(0, i-params.ContextLines)
				for j := ctxStart; j < i; j++ {
					hunk = append(hunk, edit{kind: "eq", line: edits[j].line})
				}
			}
			hunk = append(hunk, edits[i])
			i++
		}
	}
	flushHunk()

	diffOutput := result.String()

	data, _ := json.Marshal(map[string]any{
		"has_diff":  len(edits) > 0 && (stats.Additions > 0 || stats.Deletions > 0),
		"diff":      diffOutput,
		"additions": stats.Additions,
		"deletions": stats.Deletions,
		"hunks":     stats.Hunks,
	})
	return string(data), nil
}

// computeDiff 使用 Myers diff 算法的简化版本
func computeDiff(a, b []string) []edit {
	m, n := len(a), len(b)

	// 使用最长公共子序列（LCS）方法
	lcs := lcsTable(a, b)
	return backtrack(lcs, a, b, m, n)
}

func lcsTable(a, b []string) [][]int {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}
	return dp
}

func backtrack(dp [][]int, a, b []string, i, j int) []edit {
	var edits []edit
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && a[i-1] == b[j-1] {
			edits = append(edits, edit{kind: "eq", line: a[i-1]})
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			edits = append(edits, edit{kind: "ins", line: b[j-1]})
			j--
		} else if i > 0 {
			edits = append(edits, edit{kind: "del", line: a[i-1]})
			i--
		}
	}

	// 反转
	for l, r := 0, len(edits)-1; l < r; l, r = l+1, r-1 {
		edits[l], edits[r] = edits[r], edits[l]
	}
	return edits
}
