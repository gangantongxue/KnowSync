package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gangantongxue/knowsync/ks-proto/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
)

// Client 外部服务客户端
type Client struct {
	repoClient  pb.RepoServiceClient
	repoConn    *grpc.ClientConn
	gatewayAddr string
	httpClient  *http.Client
}

// NewClient 创建外部服务客户端
func NewClient(cfg *model.Config) (*Client, error) {
	conn, err := grpc.NewClient(
		cfg.RepoServer.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("连接 repo-server 失败: %w", err)
	}

	return &Client{
		repoClient:  pb.NewRepoServiceClient(conn),
		repoConn:    conn,
		gatewayAddr: cfg.Gateway.Addr,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GetNode 从 repo-server 获取节点信息（含 file_path）
func (c *Client) GetNode(ctx context.Context, userID, repoID, nodeID string) (*pb.Node, error) {
	resp, err := c.repoClient.GetNode(ctx, &pb.GetNodeRequest{
		RepoId: repoID,
		NodeId: nodeID,
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("调用 repo-server GetNode 失败: %w", err)
	}
	if resp == nil || !resp.Success {
		msg := "获取节点失败"
		if resp != nil {
			msg = resp.Msg
		}
		return nil, errors.New(msg)
	}

	return resp.Node, nil
}

// GetArticleContent 从 gateway 获取文章内容
func (c *Client) GetArticleContent(ctx context.Context, filePath string) (string, error) {
	u, err := url.Parse(c.gatewayAddr + "/internal/file")
	if err != nil {
		return "", fmt.Errorf("解析 gateway 地址失败: %w", err)
	}
	q := u.Query()
	q.Set("path", filePath)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 gateway 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gateway 返回错误状态 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 gateway 响应失败: %w", err)
	}

	slog.Debug("获取文章内容成功", "file_path", filePath, "size", len(body))
	return string(body), nil
}

// ListUserRepos 获取用户拥有的仓库 ID 列表
func (c *Client) ListUserRepos(ctx context.Context, userID string) ([]string, error) {
	resp, err := c.repoClient.ListUserRepos(ctx, &pb.ListUserReposRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("调用 repo-server ListUserRepos 失败: %w", err)
	}
	if resp == nil || !resp.Success {
		return nil, errors.New("获取用户仓库列表失败")
	}

	repoIDs := make([]string, len(resp.Repos))
	for i, repo := range resp.Repos {
		repoIDs[i] = repo.Id
	}
	return repoIDs, nil
}

// Close 关闭外部服务连接
func (c *Client) Close() error {
	return c.repoConn.Close()
}
