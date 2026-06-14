package logging

import (
	"os"
	"time"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	gray    = "\033[90m"
	white   = "\033[97m"

	slowRequest     = 100 * time.Millisecond
	verySlowRequest = 500 * time.Millisecond
)

func colorsEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

func paint(color, text string) string {
	if !colorsEnabled() {
		return text
	}
	return color + text + reset
}

func levelColor(level string) string {
	switch level {
	case "DEBUG":
		return gray
	case "INFO":
		return cyan
	case "WARN":
		return yellow
	case "ERROR":
		return red
	default:
		return white
	}
}

func statusColor(status int) string {
	switch {
	case status >= 500:
		return red
	case status >= 400:
		return yellow
	case status >= 300:
		return blue
	default:
		return green
	}
}

func durationColor(d time.Duration) string {
	switch {
	case d >= verySlowRequest:
		return red
	case d >= slowRequest:
		return yellow
	default:
		return green
	}
}
