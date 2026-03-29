package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

var (
	Default *slog.Logger
	Level   = new(slog.LevelVar)
)

func init() {
	Default = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: Level,
	}))
	slog.SetDefault(Default)
}

func ParseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
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

func SetLevel(s string) {
	Level.Set(ParseLevel(s))
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return Default
}

type loggerKey struct{}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}
