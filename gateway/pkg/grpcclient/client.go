package grpcclient

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client 管理多个后端的 gRPC 连接
// 后续添加新服务只需在配置中增加 entry，并在使用方通过 GetConn 获取连接
type Client struct {
	conns map[string]*grpc.ClientConn
}

// NewClient 根据 targets 建立 gRPC 连接，targets 为服务名到地址的映射
func NewClient(targets map[string]string) (*Client, error) {
	conns := make(map[string]*grpc.ClientConn, len(targets))

	for name, target := range targets {
		conn, err := grpc.NewClient(target,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			// 关闭已成功建立的连接
			for _, c := range conns {
				c.Close()
			}
			return nil, fmt.Errorf("连接 %s gRPC 失败: %w", name, err)
		}
		slog.Info("gRPC 客户端连接成功", "service", name, "target", target)
		conns[name] = conn
	}

	return &Client{conns: conns}, nil
}

// GetConn 按服务名获取 gRPC 连接，后续有新服务时通过此方法获取对应连接
func (c *Client) GetConn(name string) *grpc.ClientConn {
	return c.conns[name]
}

// Close 关闭所有 gRPC 连接
func (c *Client) Close() error {
	for name, conn := range c.conns {
		if err := conn.Close(); err != nil {
			slog.Error("关闭 gRPC 连接失败", "service", name, "error", err)
		}
	}
	return nil
}
