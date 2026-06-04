// Package main is the entry point for the gateway service.
package main

import (
	"log/slog"
	"os"

	"github.com/gangantongxue/knowsync/gateway/internal/app"
)

func main() {
	if err := app.NewApp(); err != nil {
		slog.Error("应用启动失败", "error", err)
		os.Exit(1)
	}
}
