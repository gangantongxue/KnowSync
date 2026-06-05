// Package logger 提供日志记录功能.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gangantongxue/knowsync/ai-server/pkg/config/model"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志记录器封装.
type Logger struct {
	Logger *slog.Logger
	Cfg    *model.Config
}

// NewLogger 创建一个新的日志记录器.
func NewLogger(cfg *model.Config) (*Logger, error) {
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Log.Dir, "ai-server.log"),
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
		LocalTime:  cfg.Log.LocalTime,
	}

	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     cfg.Log.GetLevel(),
		AddSource: true,
	})

	fileHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{
		Level:     cfg.Log.GetLevel(),
		AddSource: true,
	})

	multiHandler := newMultiHandler(consoleHandler, fileHandler)
	logger := slog.New(multiHandler)

	logger = logger.With(
		slog.String("project", "ai-server"),
	)

	slog.SetDefault(logger)

	return &Logger{Logger: logger, Cfg: cfg}, nil
}

// Printf 格式化输出日志，Redis 日志适配方法.
func (l *Logger) Printf(ctx context.Context, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	if strings.Contains(msg, "ping") {
		return
	}
	l.Logger.InfoContext(ctx, msg)
}

// MultiHandler 实现同时向多个 Handler 输出.
type MultiHandler struct {
	handlers []slog.Handler
}

// newMultiHandler 创建一个新的 MultiHandler.
func newMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// Enabled 检查是否有任一 handler 启用了该日志级别.
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle 向所有启用的 handler 分发日志记录.
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

// WithAttrs 创建一个包含指定属性的新 handler.
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

// WithGroup 创建一个包含指定分组的新 handler.
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
