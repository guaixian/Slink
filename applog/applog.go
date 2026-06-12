// Package applog 提供统一的应用日志（log/slog，输出到 stdout）。
package applog

import (
	"log/slog"
	"os"
	"strings"
)

// Logger 供业务代码直接使用
var Logger *slog.Logger

// Init 在 main 中尽早调用。可通过环境变量 SLINK_LOG_LEVEL=debug|info|warn|error 控制级别。
func Init() {
	level := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SLINK_LOG_LEVEL"))) {
	case "debug":
		level = slog.LevelDebug
	case "info", "":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	Logger = slog.New(h)
	slog.SetDefault(Logger)
	Logger.Info("applog initialized", "level", level.String())
}
