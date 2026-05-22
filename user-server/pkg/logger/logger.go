package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	Logger *slog.Logger
	Cfg    *config.Config
}

// NewLogger 创建一个新的日志记录器
func NewLogger(cfg *config.Config) *Logger {
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Logger.Dir, "user-server.log"),
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

	logger := slog.New(newMultiHandler(consoleHandler, fileHandler))

	logger = logger.With(
		slog.String("project", cfg.Project.Name),
	)

	slog.SetDefault(logger)

	return &Logger{Logger: logger, Cfg: cfg}
}

// MultiHandler 实现同时向多个 Handler 输出
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
