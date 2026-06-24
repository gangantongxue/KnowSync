package service

import (
	"bytes"
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

// Client 外部服务客户端，所有请求均通过 Gateway 的 /internal/* 接口.
type Client struct {
	gatewayAddr    string
	internalSecret string
	httpClient     *http.Client
}

// NewClient 创建外部服务客户端.
func NewClient(cfg *model.Config) (*Client, error) {
	return &Client{
		gatewayAddr:    cfg.Gateway.Addr,
		internalSecret: cfg.ServiceToken.Secret,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// doGet 向 gateway 的内部端点发送 GET 请求，自动附加 service token 和共享密钥.
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
	} else {
		req.Header.Set("X-Internal-Secret", c.internalSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 gateway 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gateway 返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// doPost 向 gateway 的内部端点发送 POST 请求.
func (c *Client) doPost(ctx context.Context, path string, query url.Values, body []byte) ([]byte, error) {
	return c.doBody(ctx, http.MethodPost, path, query, body)
}

// doPut 向 gateway 的内部端点发送 PUT 请求.
func (c *Client) doPut(ctx context.Context, path string, query url.Values, body []byte) ([]byte, error) {
	return c.doBody(ctx, http.MethodPut, path, query, body)
}

// doDelete 向 gateway 的内部端点发送 DELETE 请求.
func (c *Client) doDelete(ctx context.Context, path string, query url.Values) ([]byte, error) {
	return c.doBody(ctx, http.MethodDelete, path, query, nil)
}

// doBody 向 gateway 的内部端点发送带 body 的请求，自动附加 service token 和共享密钥.
func (c *Client) doBody(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, error) {
	serviceToken, _ := ctx.Value(tool.CtxKeyServiceToken).(string)

	u, err := url.Parse(c.gatewayAddr + path)
	if err != nil {
		return nil, fmt.Errorf("解析 gateway 地址失败: %w", err)
	}
	u.RawQuery = query.Encode()

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	if serviceToken != "" {
		req.Header.Set("Authorization", "Bearer "+serviceToken)
	} else {
		req.Header.Set("X-Internal-Secret", c.internalSecret)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 gateway 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gateway 返回错误状态 %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

// ==================== 仓库相关 ====================

// ListPublicRepos 获取所有公开仓库 ID 列表.
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

// ListUserRepos 获取用户拥有的仓库 ID 列表.
//
//nolint:revive // userID required by interface
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

// ListUserReposDetail 获取用户仓库列表（含完整信息）.
//
//nolint:dupl,revive // similar but different call; userID required by interface
func (c *Client) ListUserReposDetail(ctx context.Context, userID string) ([]tool.RepoInfo, error) {
	body, err := c.doGet(ctx, "/internal/repos/user", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repos []struct {
			ID            string `json:"id"`
			OwnerID       string `json:"owner_id"`
			Name          string `json:"name"`
			Visibility    string `json:"visibility"`
			Description   string `json:"description"`
			ArticleCount  int64  `json:"article_count"`
			FollowerCount int64  `json:"follower_count"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	infos := make([]tool.RepoInfo, 0, len(result.Repos))
	for _, r := range result.Repos {
		infos = append(infos, tool.RepoInfo{
			ID:            r.ID,
			OwnerID:       r.OwnerID,
			Name:          r.Name,
			Visibility:    r.Visibility,
			Description:   r.Description,
			ArticleCount:  r.ArticleCount,
			FollowerCount: r.FollowerCount,
		})
	}
	return infos, nil
}

// GetRepo 获取仓库详情.
//
//nolint:revive // userID required by interface
func (c *Client) GetRepo(ctx context.Context, repoID, userID string) (*tool.RepoInfo, error) {
	q := url.Values{}
	q.Set("repo_id", repoID)
	body, err := c.doGet(ctx, "/internal/repos/detail", q)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repo struct {
			ID            string `json:"id"`
			OwnerID       string `json:"owner_id"`
			Name          string `json:"name"`
			Visibility    string `json:"visibility"`
			Description   string `json:"description"`
			ArticleCount  int64  `json:"article_count"`
			FollowerCount int64  `json:"follower_count"`
		} `json:"repo"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &tool.RepoInfo{
		ID:            result.Repo.ID,
		OwnerID:       result.Repo.OwnerID,
		Name:          result.Repo.Name,
		Visibility:    result.Repo.Visibility,
		Description:   result.Repo.Description,
		ArticleCount:  result.Repo.ArticleCount,
		FollowerCount: result.Repo.FollowerCount,
	}, nil
}

// ListPublicReposDetail 获取所有公开仓库列表（含完整信息）.
//
//nolint:dupl // 与其他列表方法结构相似但调用不同接口
func (c *Client) ListPublicReposDetail(ctx context.Context) ([]tool.RepoInfo, error) {
	body, err := c.doGet(ctx, "/internal/repos/public", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repos []struct {
			ID            string `json:"id"`
			OwnerID       string `json:"owner_id"`
			Name          string `json:"name"`
			Visibility    string `json:"visibility"`
			Description   string `json:"description"`
			ArticleCount  int64  `json:"article_count"`
			FollowerCount int64  `json:"follower_count"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	infos := make([]tool.RepoInfo, 0, len(result.Repos))
	for _, r := range result.Repos {
		infos = append(infos, tool.RepoInfo{
			ID:            r.ID,
			OwnerID:       r.OwnerID,
			Name:          r.Name,
			Visibility:    r.Visibility,
			Description:   r.Description,
			ArticleCount:  r.ArticleCount,
			FollowerCount: r.FollowerCount,
		})
	}
	return infos, nil
}

// GetRepoDetail 获取仓库详情（含角色和关注状态）.
//
//nolint:revive // userID required by interface
func (c *Client) GetRepoDetail(ctx context.Context, repoID, userID string) (*tool.RepoDetail, error) {
	q := url.Values{}
	q.Set("repo_id", repoID)
	body, err := c.doGet(ctx, "/internal/repos/detail", q)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repo struct {
			ID            string `json:"id"`
			OwnerID       string `json:"owner_id"`
			Name          string `json:"name"`
			Visibility    string `json:"visibility"`
			Description   string `json:"description"`
			ArticleCount  int64  `json:"article_count"`
			FollowerCount int64  `json:"follower_count"`
		} `json:"repo"`
		MyRole      string `json:"my_role"`
		IsFollowing bool   `json:"is_following"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &tool.RepoDetail{
		RepoInfo: tool.RepoInfo{
			ID:           result.Repo.ID,
			OwnerID:      result.Repo.OwnerID,
			Name:         result.Repo.Name,
			Visibility:   result.Repo.Visibility,
			Description:  result.Repo.Description,
			ArticleCount: result.Repo.ArticleCount,
		},
		FollowerCount: result.Repo.FollowerCount,
		MyRole:        result.MyRole,
		IsFollowing:   result.IsFollowing,
	}, nil
}

// ==================== 文件相关 ====================

// GetArticleContent 从 gateway 获取文章内容.
func (c *Client) GetArticleContent(ctx context.Context, userID, repoID, filePath string) (string, error) {
	q := url.Values{}
	q.Set("path", userID+"/"+repoID+"/"+filePath)
	body, err := c.doGet(ctx, "/internal/file", q)
	if err != nil {
		return "", err
	}
	slog.Debug("获取文章内容成功", "file_path", filePath, "size", len(body))
	return string(body), nil
}

// GetFileContent 获取仓库文件内容（实现 tool.FileClient 接口）.
func (c *Client) GetFileContent(ctx context.Context, ownerID, repoID, filePath string) (string, error) {
	return c.GetArticleContent(ctx, ownerID, repoID, filePath)
}

// ListRepoFiles 获取仓库文件列表.
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

// ==================== 文件写入 ====================

// CreateFile 创建文件.
func (c *Client) CreateFile(ctx context.Context, ownerID, repoID, filePath, content string) error {
	q := url.Values{}
	q.Set("owner_id", ownerID)
	q.Set("repo_id", repoID)
	q.Set("path", filePath)
	body, _ := json.Marshal(map[string]string{"content": content})
	_, err := c.doPost(ctx, "/internal/repos/files", q, body)
	return err
}

// UpdateFile 更新文件内容.
func (c *Client) UpdateFile(ctx context.Context, ownerID, repoID, filePath, content string) error {
	q := url.Values{}
	q.Set("owner_id", ownerID)
	q.Set("repo_id", repoID)
	q.Set("path", filePath)
	body, _ := json.Marshal(map[string]string{"content": content})
	_, err := c.doPut(ctx, "/internal/repos/files", q, body)
	return err
}

// DeleteFile 删除文件.
func (c *Client) DeleteFile(ctx context.Context, ownerID, repoID, filePath string) error {
	q := url.Values{}
	q.Set("owner_id", ownerID)
	q.Set("repo_id", repoID)
	q.Set("path", filePath)
	_, err := c.doDelete(ctx, "/internal/repos/files", q)
	return err
}

// RenameFile 重命名/移动文件.
func (c *Client) RenameFile(ctx context.Context, ownerID, repoID, oldPath, newPath string) error {
	q := url.Values{}
	q.Set("owner_id", ownerID)
	q.Set("repo_id", repoID)
	q.Set("path", oldPath)
	body, _ := json.Marshal(map[string]string{"new_path": newPath})
	_, err := c.doPut(ctx, "/internal/repos/files/rename", q, body)
	return err
}

// ==================== 知识库写入 ====================

// CreateRepo 创建知识库.
//
//nolint:revive // userID required by interface
func (c *Client) CreateRepo(ctx context.Context, userID, name, description, visibility string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"name":        name,
		"description": description,
		"visibility":  visibility,
	})
	resp, err := c.doPost(ctx, "/internal/repos", nil, body)
	if err != nil {
		return "", err
	}
	var result struct {
		RepoID string `json:"repo_id"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	return result.RepoID, nil
}

// UpdateRepo 更新知识库.
//
//nolint:revive // userID required by interface
func (c *Client) UpdateRepo(ctx context.Context, repoID, userID, name, description, visibility string) error {
	body, _ := json.Marshal(map[string]string{
		"name":        name,
		"description": description,
		"visibility":  visibility,
	})
	q := url.Values{}
	q.Set("repo_id", repoID)
	_, err := c.doPut(ctx, "/internal/repos", q, body)
	return err
}

// ==================== 用户搜索 ====================

// SearchUsers 搜索用户.
func (c *Client) SearchUsers(ctx context.Context, keyword string) ([]tool.UserInfo, error) {
	q := url.Values{}
	q.Set("q", keyword)
	body, err := c.doGet(ctx, "/internal/users/search", q)
	if err != nil {
		return nil, err
	}
	var result struct {
		Users []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"users"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	users := make([]tool.UserInfo, 0, len(result.Users))
	for _, u := range result.Users {
		users = append(users, tool.UserInfo{ID: u.ID, Name: u.Name})
	}
	return users, nil
}

// ==================== 关注操作 ====================

// FollowRepo 关注知识库.
//
//nolint:revive // userID required by interface
func (c *Client) FollowRepo(ctx context.Context, userID, repoID string) error {
	q := url.Values{}
	q.Set("repo_id", repoID)
	_, err := c.doPost(ctx, "/internal/repos/follow", q, nil)
	return err
}

// UnfollowRepo 取消关注知识库.
//
//nolint:revive // userID required by interface
func (c *Client) UnfollowRepo(ctx context.Context, userID, repoID string) error {
	q := url.Values{}
	q.Set("repo_id", repoID)
	_, err := c.doDelete(ctx, "/internal/repos/follow", q)
	return err
}

// ListFollowedRepos 获取关注的知识库列表（含完整信息）.
//
//nolint:dupl,revive // similar but different call; userID required by interface
func (c *Client) ListFollowedRepos(ctx context.Context, userID string) ([]tool.RepoInfo, error) {
	body, err := c.doGet(ctx, "/internal/repos/followed", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repos []struct {
			ID            string `json:"id"`
			OwnerID       string `json:"owner_id"`
			Name          string `json:"name"`
			Visibility    string `json:"visibility"`
			Description   string `json:"description"`
			ArticleCount  int64  `json:"article_count"`
			FollowerCount int64  `json:"follower_count"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	infos := make([]tool.RepoInfo, 0, len(result.Repos))
	for _, r := range result.Repos {
		infos = append(infos, tool.RepoInfo{
			ID:            r.ID,
			OwnerID:       r.OwnerID,
			Name:          r.Name,
			Visibility:    r.Visibility,
			Description:   r.Description,
			ArticleCount:  r.ArticleCount,
			FollowerCount: r.FollowerCount,
		})
	}
	return infos, nil
}

// ==================== 协作者管理 ====================

// AddCollaborator 添加协作者.
func (c *Client) AddCollaborator(ctx context.Context, repoID, userID, role string) error {
	q := url.Values{}
	q.Set("repo_id", repoID)
	body, _ := json.Marshal(map[string]string{
		"user_id": userID,
		"role":    role,
	})
	_, err := c.doPost(ctx, "/internal/repos/collaborators", q, body)
	return err
}

// RemoveCollaborator 移除协作者.
func (c *Client) RemoveCollaborator(ctx context.Context, repoID, userID string) error {
	q := url.Values{}
	q.Set("repo_id", repoID)
	q.Set("user_id", userID)
	_, err := c.doDelete(ctx, "/internal/repos/collaborators", q)
	return err
}

// UpdateCollaboratorRole 更新协作者角色.
func (c *Client) UpdateCollaboratorRole(ctx context.Context, repoID, userID, role string) error {
	q := url.Values{}
	q.Set("repo_id", repoID)
	body, _ := json.Marshal(map[string]string{
		"user_id": userID,
		"role":    role,
	})
	_, err := c.doPut(ctx, "/internal/repos/collaborators/role", q, body)
	return err
}

// ListCollaborators 列出协作者.
func (c *Client) ListCollaborators(ctx context.Context, repoID string) ([]tool.CollaboratorInfo, error) {
	q := url.Values{}
	q.Set("repo_id", repoID)
	body, err := c.doGet(ctx, "/internal/repos/collaborators", q)
	if err != nil {
		return nil, err
	}
	var result struct {
		Collaborators []struct {
			UserID   string `json:"user_id"`
			UserName string `json:"user_name"`
			Role     string `json:"role"`
		} `json:"collaborators"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	items := make([]tool.CollaboratorInfo, 0, len(result.Collaborators))
	for _, c := range result.Collaborators {
		items = append(items, tool.CollaboratorInfo{
			UserID:   c.UserID,
			UserName: c.UserName,
			Role:     c.Role,
		})
	}
	return items, nil
}

// Close 关闭外部服务连接.
func (c *Client) Close() error {
	return nil
}
