package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gangantongxue/knowsync/gateway/pkg/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 封装 slog.Logger，支持同时向控制台和文件输出
type Logger struct {
	Logger       *slog.Logger
	MultiHandler *MultiHandler
	Cfg          *config.Config
}

// NewLogger 创建日志记录器，控制台使用文本格式，文件使用 JSON 格式 + 自动轮转
func NewLogger(cfg *config.Config) (*Logger, error) {
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Logger.Dir, "gateway.log"),
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
		LocalTime:  cfg.Logger.LocalTime,
	}

	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     cfg.Logger.GetLevel(),
		AddSource: true,
	})

	fileHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{
		Level:     cfg.Logger.GetLevel(),
		AddSource: true,
	})

	multiHandler := newMultiHandler(consoleHandler, fileHandler)
	logger := slog.New(multiHandler)

	logger = logger.With(
		slog.String("project", cfg.Project.Name),
	)

	slog.SetDefault(logger)

	return &Logger{Logger: logger, MultiHandler: multiHandler, Cfg: cfg}, nil
}

// Printf 格式化输出日志，用于 Redis 等第三方库的日志适配
func (l *Logger) Printf(ctx context.Context, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	if strings.Contains(msg, "ping") {
		return
	}
	l.Logger.InfoContext(ctx, msg)
}

// MultiHandler 实现同时向多个 slog.Handler 输出
type MultiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

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

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}
