package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gangantongxue/knowsync/ai-server/internal/llm/tool"
	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

// Client 外部服务客户端，所有请求均通过 Gateway 的 /internal/* 接口
type Client struct {
	gatewayAddr string
	httpClient  *http.Client
}

// NewClient 创建外部服务客户端
func NewClient(cfg *model.Config) (*Client, error) {
	return &Client{
		gatewayAddr: cfg.Gateway.Addr,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// doGet 向 gateway 的内部端点发送 GET 请求，自动附加 service token
func (c *Client) doGet(ctx context.Context, path string, query url.Values) ([]byte, error) {
	serviceToken, _ := ctx.Value(tool.CtxKeyServiceToken).(string)

	u, err := url.Parse(c.gatewayAddr + path)
	if err != nil {
		return nil, fmt.Errorf("解析 gateway 地址失败: %w", err)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	if serviceToken != "" {
		req.Header.Set("Authorization", "Bearer "+serviceToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 gateway 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gateway 返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// ==================== 仓库相关 ====================

// ListPublicRepos 获取所有公开仓库 ID 列表
func (c *Client) ListPublicRepos(ctx context.Context) ([]string, error) {
	body, err := c.doGet(ctx, "/internal/repos/public", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repos []struct {
			ID string `json:"id"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	repoIDs := make([]string, len(result.Repos))
	for i, r := range result.Repos {
		repoIDs[i] = r.ID
	}
	return repoIDs, nil
}

// ListUserRepos 获取用户拥有的仓库 ID 列表
func (c *Client) ListUserRepos(ctx context.Context, userID string) ([]string, error) {
	body, err := c.doGet(ctx, "/internal/repos/user", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repos []struct {
			ID string `json:"id"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	repoIDs := make([]string, len(result.Repos))
	for i, r := range result.Repos {
		repoIDs[i] = r.ID
	}
	return repoIDs, nil
}

// ListUserReposDetail 获取用户仓库列表（含完整信息）
func (c *Client) ListUserReposDetail(ctx context.Context, userID string) ([]tool.RepoInfo, error) {
	body, err := c.doGet(ctx, "/internal/repos/user", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repos []struct {
			ID           string `json:"id"`
			OwnerID      string `json:"owner_id"`
			Name         string `json:"name"`
			Visibility   string `json:"visibility"`
			Description  string `json:"description"`
			ArticleCount int64  `json:"article_count"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	infos := make([]tool.RepoInfo, 0, len(result.Repos))
	for _, r := range result.Repos {
		infos = append(infos, tool.RepoInfo{
			ID:           r.ID,
			OwnerID:      r.OwnerID,
			Name:         r.Name,
			Visibility:   r.Visibility,
			Description:  r.Description,
			ArticleCount: r.ArticleCount,
		})
	}
	return infos, nil
}

// GetRepo 获取仓库详情
func (c *Client) GetRepo(ctx context.Context, repoID, userID string) (*tool.RepoInfo, error) {
	q := url.Values{}
	q.Set("repo_id", repoID)
	body, err := c.doGet(ctx, "/internal/repos/detail", q)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repo struct {
			ID           string `json:"id"`
			OwnerID      string `json:"owner_id"`
			Name         string `json:"name"`
			Visibility   string `json:"visibility"`
			Description  string `json:"description"`
			ArticleCount int64  `json:"article_count"`
		} `json:"repo"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &tool.RepoInfo{
		ID:           result.Repo.ID,
		OwnerID:      result.Repo.OwnerID,
		Name:         result.Repo.Name,
		Visibility:   result.Repo.Visibility,
		Description:  result.Repo.Description,
		ArticleCount: result.Repo.ArticleCount,
	}, nil
}

// ==================== 文件相关 ====================

// GetArticleContent 从 gateway 获取文章内容
func (c *Client) GetArticleContent(ctx context.Context, userID string, repoID, filePath string) (string, error) {
	q := url.Values{}
	q.Set("path", userID+"/"+repoID+"/"+filePath)
	body, err := c.doGet(ctx, "/internal/file", q)
	if err != nil {
		return "", err
	}
	slog.Debug("获取文章内容成功", "file_path", filePath, "size", len(body))
	return string(body), nil
}

// GetFileContent 获取仓库文件内容（实现 tool.FileClient 接口）
func (c *Client) GetFileContent(ctx context.Context, ownerID, repoID, filePath string) (string, error) {
	return c.GetArticleContent(ctx, ownerID, repoID, filePath)
}

// ListRepoFiles 获取仓库文件列表
func (c *Client) ListRepoFiles(ctx context.Context, ownerID, repoID, dirPath string) ([]tool.FileEntry, error) {
	q := url.Values{}
	q.Set("owner_id", ownerID)
	q.Set("repo_id", repoID)
	q.Set("path", dirPath)
	body, err := c.doGet(ctx, "/internal/repos/tree", q)
	if err != nil {
		return nil, err
	}

	var result struct {
		Entries []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			Size int64  `json:"size"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析文件列表响应失败: %w", err)
	}

	entries := make([]tool.FileEntry, 0, len(result.Entries))
	for _, e := range result.Entries {
		p := dirPath
		if p != "" {
			p += "/"
		}
		p += e.Name
		entries = append(entries, tool.FileEntry{
			Name: e.Name,
			Type: e.Type,
			Size: e.Size,
			Path: p,
		})
	}
	return entries, nil
}

// Close 关闭外部服务连接
func (c *Client) Close() error {
	return nil
}
