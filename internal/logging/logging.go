package logging

import (
	"log/slog"
	"os"
	"strings"
)

func New(level, env string) *slog.Logger {
	parsedLevel := parseLevel(level)

	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parsedLevel})
	} else {
		handler = NewColoredHandler(os.Stdout, parsedLevel)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
