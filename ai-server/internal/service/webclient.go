package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

// WebClient 外部网络服务客户端.
type WebClient struct {
	httpClient *http.Client
	searchURL  string
}

// NewWebClient 创建外部网络服务客户端.
func NewWebClient(cfg *model.WebSearchCfg) *WebClient {
	searchURL := cfg.BaseURL
	if searchURL == "" {
		searchURL = "https://api.duckduckgo.com/"
	}
	return &WebClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		searchURL:  searchURL,
	}
}

// FetchURL 获取 URL 内容.
func (w *WebClient) FetchURL(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "KnowSync/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("返回错误状态 %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Search 执行网络搜索.
//
//nolint:gocyclo // 搜索需要处理多种搜索结果格式和错误情况
func (w *WebClient) Search(ctx context.Context, query string) (string, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("no_html", "1")
	q.Set("skip_disambig", "1")

	fullURL := w.searchURL + "?" + q.Encode()
	slog.Info("执行网络搜索", "query", query, "url", fullURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建搜索请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "KnowSync/1.0")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("搜索请求失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取搜索响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("搜索服务返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	// 解析 DuckDuckGo Instant Answer 格式
	var ddgResponse struct {
		Abstract         string `json:"Abstract"`
		AbstractSource   string `json:"AbstractSource"`
		AbstractURL      string `json:"AbstractURL"`
		Answer           string `json:"Answer"`
		AnswerType       string `json:"AnswerType"`
		Infobox          string `json:"Infobox"`
		Image            string `json:"Image"`
		Definition       string `json:"Definition"`
		DefinitionSource string `json:"DefinitionSource"`
		Type             string `json:"Type"`
		RelatedTopics    []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
			Topics   []struct {
				Text     string `json:"Text"`
				FirstURL string `json:"FirstURL"`
			} `json:"Topics"`
		} `json:"RelatedTopics"`
		Results []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(body, &ddgResponse); err != nil {
		// 如果解析失败，返回原文
		return string(body), nil
	}

	// 格式化结果
	result := fmt.Sprintf("搜索结果: %s\n\n", query)

	if ddgResponse.Answer != "" {
		result += fmt.Sprintf("【直接答案】\n%s\n\n", ddgResponse.Answer)
	}

	if ddgResponse.Abstract != "" {
		result += fmt.Sprintf("【摘要】\n%s\n来源: %s\n链接: %s\n\n",
			ddgResponse.Abstract, ddgResponse.AbstractSource, ddgResponse.AbstractURL)
	}

	if ddgResponse.Definition != "" {
		result += fmt.Sprintf("【定义】\n%s\n来源: %s\n\n",
			ddgResponse.Definition, ddgResponse.DefinitionSource)
	}

	if len(ddgResponse.Results) > 0 {
		result += "【搜索结果】\n"
		var resultSb138 strings.Builder
		for i, r := range ddgResponse.Results {
			fmt.Fprintf(&resultSb138, "%d. %s\n   链接: %s\n", i+1, r.Text, r.FirstURL)
		}
		result += resultSb138.String()
		result += "\n"
	}

	if len(ddgResponse.RelatedTopics) > 0 {
		count := 0
		result += "【相关主题】\n"
		var resultSb147 strings.Builder
		var resultSb151 strings.Builder
		for _, topic := range ddgResponse.RelatedTopics {
			if count >= 10 {
				break
			}
			if topic.Text != "" {
				count++
				fmt.Fprintf(&resultSb147, "%d. %s\n   链接: %s\n", count, topic.Text, topic.FirstURL)
			}
			var resultSb155 strings.Builder
			for _, sub := range topic.Topics {
				if count >= 10 {
					break
				}
				if sub.Text != "" {
					count++
					fmt.Fprintf(&resultSb155, "%d. %s\n   链接: %s\n", count, sub.Text, sub.FirstURL)
				}
			}
			resultSb151.WriteString(resultSb155.String())
		}
		result += resultSb151.String()
		result += resultSb147.String()
		result += "\n"
	}

	if result == fmt.Sprintf("搜索结果: %s\n\n", query) {
		result += "未找到相关结果"
	}

	return result, nil
}
