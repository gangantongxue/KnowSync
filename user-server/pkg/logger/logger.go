package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gangantongxue/knowsync/user-server/pkg/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	Logger       *slog.Logger
	MultiHandler *MultiHandler
	Cfg          *config.Config
}

// NewLogger 创建一个新的日志记录器
func NewLogger(cfg *config.Config) (*Logger, error) {
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Logger.Dir, "user-server.log"),
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
		LocalTime:  cfg.Logger.LocalTime,
	}

	// 控制台输出为文本格式，包含文件名和行号
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     cfg.Logger.GetLevel(),
		AddSource: true,
	})

	// 文件志输出为 JSON 格式，包含文件名和行号
	fileHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{
		Level:     cfg.Logger.GetLevel(),
		AddSource: true,
	})

	multiHandler := newMultiHandler(consoleHandler, fileHandler)
	logger := slog.New(multiHandler)

	logger = logger.With(
		slog.String("project", cfg.Project.Name),
	)

	// 设置为默认日志记录器
	// 后续可直接使用 slog.Info() 等方法记录日志
	slog.SetDefault(logger)

	return &Logger{Logger: logger, MultiHandler: multiHandler, Cfg: cfg}, nil
}

// Printf 格式化输出日志，Redis 日志适配方法
func (l *Logger) Printf(ctx context.Context, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	// 过滤 Redis 日志中的 ping 消息
	if strings.Contains(msg, "ping") {
		return
	}
	l.Logger.InfoContext(ctx, msg)
}

// MultiHandler 实现同时向多个 Handler 输出
type MultiHandler struct {
	handlers []slog.Handler
}

// newMultiHandler 创建一个新的 MultiHandler
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
