package logger

import (
	"log/slog"
	"os"
	"strings"
)

func Init(levelStr string) {
	level := parseLevel(levelStr)
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func parseLevel(levelStr string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
