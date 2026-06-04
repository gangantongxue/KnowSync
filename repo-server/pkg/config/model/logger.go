package model

import "log/slog"

// LoggerCfg 日志配置项.
type LoggerCfg struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Dir        string `yaml:"dir" mapstructure:"dir"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`
	Compress   bool   `yaml:"compress" mapstructure:"compress"`
	LocalTime  bool   `yaml:"local_time" mapstructure:"local_time"`
}

// GetLevel 将配置中的日志级别字符串转换为 slog.Level.
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
