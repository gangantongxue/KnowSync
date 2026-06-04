package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

//nolint:revive // self-documenting
type ImportFromURL struct {
	webFetcher      WebFetcher
	fileWriteClient FileWriteClient
}

//nolint:revive // self-documenting
func NewImportFromURL(wf WebFetcher, fwc FileWriteClient) *ImportFromURL {
	return &ImportFromURL{
		webFetcher:      wf,
		fileWriteClient: fwc,
	}
}

//nolint:revive // self-documenting
func (i *ImportFromURL) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "import_from_url",
		Desc: "从指定 URL 抓取网页内容并保存为知识库中的文件。适用于将网页文章、文档等保存到知识库。注意：部分网站可能限制抓取。如果用户已明确要求，可以设置 _skip_confirm: true 跳过确认。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     TypeString,
				Desc:     "要抓取的网页 URL",
				Required: true,
			},
			ParamRepoID: {
				Type:     TypeString,
				Desc:     "目标知识库 ID",
				Required: true,
			},
			ParamFilePath: {
				Type:     TypeString,
				Desc:     "保存的文件路径（可选，默认从 URL 自动生成）",
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
func (i *ImportFromURL) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	return i.execute(ctx, arguments)
}

func (i *ImportFromURL) execute(ctx context.Context, paramsJSON string) (string, error) {
	var params struct {
		URL      string `json:"url"`
		RepoID   string `json:"repo_id"`
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %w", err)
	}

	if params.URL == "" {
		return `{"success": false, "message": "url 不能为空"}`, nil
	}
	if params.RepoID == "" {
		return ErrRespRepoIDEmpty, nil
	}

	userID, _ := ctx.Value(CtxKeyUserID).(string)
	if userID == "" {
		return ErrRespUserInfo, nil
	}

	slog.Info("从 URL 导入内容", "url", params.URL, "repo_id", params.RepoID)

	// 1. 获取 URL 内容
	body, err := i.webFetcher.FetchURL(ctx, params.URL)
	if err != nil {
		return fmt.Sprintf(`{"success": false, "message": "抓取 URL 失败: %s"}`, err.Error()), nil
	}

	// 2. 确定文件路径
	filePath := params.FilePath
	if filePath == "" {
		filePath = generateFilePath(params.URL)
	}

	// 3. 处理内容：检测是否为 HTML，尝试提取文本
	contentType := detectContentType(body)
	var content string

	switch contentType {
	case "html":
		content = extractTextFromHTML(body, params.URL)
	case "json":
		content = formatJSON(body)
	default:
		content = string(body)
	}

	if strings.TrimSpace(content) == "" {
		return `{"success": false, "message": "抓取的内容为空"}`, nil
	}

	// 4. 保存文件
	if err := i.fileWriteClient.CreateFile(ctx, userID, params.RepoID, filePath, content); err != nil {
		return fmt.Sprintf(`{"success": false, "message": "保存文件失败: %s"}`, err.Error()), nil
	}

	data, _ := json.Marshal(map[string]any{
		KeySuccess:    true,
		"url":         params.URL,
		ParamRepoID:   params.RepoID,
		ParamFilePath: filePath,
		ParamSize:     len(content),
	})
	return string(data), nil
}

func generateFilePath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "imported/page.md"
	}

	p := parsed.Path
	if p == "" || p == "/" {
		p = "/index"
	}

	ext := path.Ext(p)
	if ext == "" || ext == ".html" || ext == ".htm" {
		p = strings.TrimSuffix(p, ext) + ".md"
	}

	p = strings.TrimPrefix(p, "/")
	if p == "" {
		p = "index.md"
	}

	return path.Join("imported", p)
}

func detectContentType(body []byte) string {
	s := string(body)
	trimmed := strings.TrimSpace(s)

	if len(trimmed) == 0 {
		return "unknown"
	}

	// 检测 JSON
	if trimmed[0] == '{' || trimmed[0] == '[' {
		return "json"
	}

	// 检测 HTML
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "<!doctype html") || strings.Contains(lower, "<html") {
		return "html"
	}

	return "text"
}

//nolint:gocyclo // HTML 提取需要处理多种标签和嵌套结构
func extractTextFromHTML(body []byte, pageURL string) string {
	s := string(body)

	// 尝试提取 title
	title := ""
	if i := strings.Index(strings.ToLower(s), "<title"); i >= 0 {
		start := strings.Index(s[i:], ">") + i + 1
		end := strings.Index(s[start:], "</title>")
		if end > 0 {
			title = strings.TrimSpace(s[start : start+end])
		}
	}

	// 去除 script 和 style 标签
	for {
		start := strings.Index(s, "<script")
		if start < 0 {
			break
		}
		end := strings.Index(s[start:], "</script>")
		if end < 0 {
			break
		}
		s = s[:start] + s[start+end+9:]
	}
	for {
		start := strings.Index(s, "<style")
		if start < 0 {
			break
		}
		end := strings.Index(s[start:], "</style>")
		if end < 0 {
			break
		}
		s = s[:start] + s[start+end+7:]
	}

	// 提取 body 内容
	bodyContent := s
	if i := strings.Index(strings.ToLower(s), "<body"); i >= 0 {
		start := strings.Index(s[i:], ">") + i + 1
		end := strings.LastIndex(s, "</body>")
		if end > start {
			bodyContent = s[start:end]
		}
	}

	// 将块级标签替换为换行
	for _, tag := range []string{"</p>", "</div>", "</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>", "</li>", "</tr>", "</blockquote>", "</pre>"} {
		bodyContent = strings.ReplaceAll(bodyContent, tag, "\n")
	}
	bodyContent = strings.ReplaceAll(bodyContent, "<br>", "\n")
	bodyContent = strings.ReplaceAll(bodyContent, "<br/>", "\n")
	bodyContent = strings.ReplaceAll(bodyContent, "<br />", "\n")
	bodyContent = strings.ReplaceAll(bodyContent, "</td>", "\t")
	bodyContent = strings.ReplaceAll(bodyContent, "</th>", "\t")

	// 去除所有标签
	var result strings.Builder
	inTag := false
	for _, r := range bodyContent {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	text := result.String()

	// HTML 实体解码
	text = decodeHTMLEntities(text)

	// 合并空行
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		} else if len(cleaned) > 0 && cleaned[len(cleaned)-1] != "" {
			cleaned = append(cleaned, "")
		}
	}

	output := strings.Join(cleaned, "\n")

	// 添加标题和来源信息
	if title != "" {
		output = "# " + title + "\n\n> 来源: " + pageURL + "\n\n" + output
	} else {
		output = "> 来源: " + pageURL + "\n\n" + output
	}

	return output
}

func decodeHTMLEntities(s string) string {
	entities := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
		"&nbsp;": " ",
		"&copy;": "©",
	}
	for k, v := range entities {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

func formatJSON(body []byte) string {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return string(body)
	}
	formatted, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return string(body)
	}
	return string(formatted)
}
