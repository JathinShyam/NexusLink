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

// TextHandler writes human-readable log lines. Use colored=true for terminals,
// colored=false for log files (same layout, no ANSI escape codes).
type TextHandler struct {
	w       io.Writer
	level   slog.Level
	mu      sync.Mutex
	colored bool
}

func NewColoredHandler(w io.Writer, level slog.Level) *TextHandler {
	return &TextHandler{w: w, level: level, colored: true}
}

func NewPlainHandler(w io.Writer, level slog.Level) *TextHandler {
	return &TextHandler{w: w, level: level, colored: false}
}

func (h *TextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *TextHandler) Handle(_ context.Context, r slog.Record) error {
	var buf strings.Builder

	ts := r.Time.Format("15:04:05")
	level := r.Level.String()

	buf.WriteString(h.styled(dim, ts))
	buf.WriteByte(' ')
	buf.WriteString(h.styled(levelColor(level), fmt.Sprintf("%-5s", level)))

	if r.Message == httpRequestMsg {
		h.formatHTTPRequest(&buf, r)
	} else {
		buf.WriteByte(' ')
		buf.WriteString(h.styled(bold+white, r.Message))
		h.appendAttrs(&buf, r)
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, buf.String())
	return err
}

func (h *TextHandler) styled(color, text string) string {
	if !h.colored {
		return text
	}
	return paint(color, text)
}

func (h *TextHandler) formatHTTPRequest(buf *strings.Builder, r slog.Record) {
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
	buf.WriteString(h.styled(bold+cyan, method))
	buf.WriteByte(' ')
	buf.WriteString(h.styled(white, path))
	buf.WriteString(h.styled(dim, " → "))
	buf.WriteString(h.styled(statusColor(status), fmt.Sprintf("%d", status)))
	buf.WriteString(h.styled(dim, " in "))
	buf.WriteString(h.styled(durationColor(duration), formatDuration(duration)))
}

func (h *TextHandler) appendAttrs(buf *strings.Builder, r slog.Record) {
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(' ')
		buf.WriteString(h.styled(dim, a.Key+"="))
		buf.WriteString(h.formatAttrValue(a.Value))
		return true
	})
}

func (h *TextHandler) formatAttrValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return h.styled(white, v.String())
	case slog.KindInt64:
		return h.styled(magenta, fmt.Sprintf("%d", v.Int64()))
	case slog.KindDuration:
		return h.styled(green, v.Duration().String())
	case slog.KindAny:
		if err, ok := v.Any().(error); ok {
			return h.styled(red, err.Error())
		}
		return h.styled(white, fmt.Sprintf("%v", v.Any()))
	default:
		return h.styled(white, v.String())
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

func (h *TextHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *TextHandler) WithGroup(_ string) slog.Handler {
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
