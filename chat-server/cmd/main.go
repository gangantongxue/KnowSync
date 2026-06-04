package main

import (
	"log/slog"
	"os"

	"github.com/gangantongxue/knowsync/chat-server/internal/app"
)

func main() {
	if err := app.NewApp(); err != nil {
		slog.Error("应用启动失败", "error", err)
		os.Exit(1)
	}
}
