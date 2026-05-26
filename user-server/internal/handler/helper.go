package handler

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// getClientIP 获取客户端 IP 地址
func getClientIP(ctx context.Context) string {
	// 从元数据中获取 IP 地址
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-forwarded-for"); len(vals) > 0 && vals[0] != "" {
			return strings.TrimSpace(strings.Split(vals[0], ",")[0])
		}
		if vals := md.Get("x-real-ip"); len(vals) > 0 && vals[0] != "" {
			return strings.TrimSpace(vals[0])
		}
	}
	// 从上下文获取 IP 地址
	if p, ok := peer.FromContext(ctx); ok {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err != nil {
			return p.Addr.String()
		}
		return host
	}
	return ""
}
