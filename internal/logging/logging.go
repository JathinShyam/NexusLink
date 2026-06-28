package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Options struct {
	Level   string
	Env     string
	LogDir  string
	LogFile string
}

// Setup creates a logger that writes to stdout and an append-only file in LogDir.
// The returned closer must be called on shutdown (e.g. defer closer.Close()).
func Setup(opts Options) (*slog.Logger, io.Closer, error) {
	parsedLevel := parseLevel(opts.Level)

	logDir := opts.LogDir
	if logDir == "" {
		logDir = "logs/api"
	}
	logFile := opts.LogFile
	if logFile == "" {
		logFile = "app.log"
	}

	file, err := openLogFile(logDir, logFile)
	if err != nil {
		return nil, nil, err
	}

	var fileHandler slog.Handler = NewPlainHandler(file, parsedLevel)

	var consoleHandler slog.Handler
	if opts.Env == "production" {
		consoleHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parsedLevel})
		fileHandler = slog.NewJSONHandler(file, &slog.HandlerOptions{Level: parsedLevel})
	} else {
		consoleHandler = NewColoredHandler(os.Stdout, parsedLevel)
	}

	logger := slog.New(NewMultiHandler(consoleHandler, fileHandler))
	return logger, file, nil
}

// New is kept for tests and simple callers; logs to stdout only.
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

// CloseCloser closes an io.Closer if non-nil.
func CloseCloser(c io.Closer) error {
	if c == nil {
		return nil
	}
	if err := c.Close(); err != nil {
		return fmt.Errorf("close log file: %w", err)
	}
	return nil
}
