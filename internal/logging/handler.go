package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const httpRequestMsg = "http.request"

type ColoredHandler struct {
	w     io.Writer
	level slog.Level
	mu    sync.Mutex
}

func NewColoredHandler(w io.Writer, level slog.Level) *ColoredHandler {
	return &ColoredHandler{w: w, level: level}
}

func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	var buf strings.Builder

	ts := r.Time.Format("15:04:05")
	level := r.Level.String()

	buf.WriteString(paint(dim, ts))
	buf.WriteByte(' ')
	buf.WriteString(paint(levelColor(level), fmt.Sprintf("%-5s", level)))

	if r.Message == httpRequestMsg {
		h.formatHTTPRequest(&buf, r)
	} else {
		buf.WriteByte(' ')
		buf.WriteString(paint(bold+white, r.Message))
		h.appendAttrs(&buf, r)
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, buf.String())
	return err
}

func (h *ColoredHandler) formatHTTPRequest(buf *strings.Builder, r slog.Record) {
	var method, path string
	var status int
	var duration time.Duration

	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "method":
			method, _ = a.Value.Any().(string)
		case "path":
			path, _ = a.Value.Any().(string)
		case "status":
			switch v := a.Value.Any().(type) {
			case int:
				status = v
			case int64:
				status = int(v)
			}
		case "duration":
			duration, _ = a.Value.Any().(time.Duration)
		}
		return true
	})

	buf.WriteByte(' ')
	buf.WriteString(paint(bold+cyan, method))
	buf.WriteByte(' ')
	buf.WriteString(paint(white, path))
	buf.WriteString(paint(dim, " → "))
	buf.WriteString(paint(statusColor(status), fmt.Sprintf("%d", status)))
	buf.WriteString(paint(dim, " in "))
	buf.WriteString(paint(durationColor(duration), formatDuration(duration)))
}

func (h *ColoredHandler) appendAttrs(buf *strings.Builder, r slog.Record) {
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(' ')
		buf.WriteString(paint(dim, a.Key+"="))
		buf.WriteString(formatAttrValue(a.Value))
		return true
	})
}

func formatAttrValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return paint(white, v.String())
	case slog.KindInt64:
		return paint(magenta, fmt.Sprintf("%d", v.Int64()))
	case slog.KindDuration:
		return paint(green, v.Duration().String())
	case slog.KindAny:
		if err, ok := v.Any().(error); ok {
			return paint(red, err.Error())
		}
		return paint(white, fmt.Sprintf("%v", v.Any()))
	default:
		return paint(white, v.String())
	}
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	case d < time.Second:
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000)
	default:
		return d.Round(time.Millisecond).String()
	}
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColoredHandler) WithGroup(_ string) slog.Handler {
	return h
}

// LogHTTPRequest logs a formatted HTTP request line in development.
func LogHTTPRequest(log *slog.Logger, method, path string, status int, duration time.Duration) {
	log.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		httpRequestMsg,
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Any("duration", duration),
	)
}
