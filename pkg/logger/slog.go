package logger

import (
	"log/slog"
	"os"
	"strings"
)

func Setup(level string) *slog.Logger {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
		slog.Warn("Invalid log level specified, using default level: info", slog.String("invalid_level", level))
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	logger := slog.New(handler)

	logger.Info("Logger initialized", slog.String("level", logLevel.String()))

	return logger
}
