package model

import "log/slog"

// LoggerCfg 日志配置项
type LoggerCfg struct {
	Level      string `yaml:"level"`
	Dir        string `yaml:"dir"`         // Dir 日志目录
	MaxSize    int    `yaml:"max_size"`    // MaxSize 日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"` // MaxBackups 日志文件最大备份数量
	MaxAge     int    `yaml:"max_age"`     // MaxAge 日志文件最大年龄（天）
	Compress   bool   `yaml:"compress"`    // Compress 是否压缩日志文件
	LocalTime  bool   `yaml:"local_time"`  // LocalTime 是否使用本地时间
}

func (l *LoggerCfg) GetLevel() slog.Level {
	switch l.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
