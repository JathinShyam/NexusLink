package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestColoredHandlerHTTPRequest(t *testing.T) {
	var buf bytes.Buffer
	handler := NewColoredHandler(&buf, parseLevel("info"))

	record := slog.NewRecord(
		time.Date(2026, 6, 13, 19, 12, 3, 0, time.UTC),
		slog.LevelInfo,
		httpRequestMsg,
		0,
	)
	record.AddAttrs(
		slog.String("method", "GET"),
		slog.String("path", "/health"),
		slog.Int("status", 200),
		slog.Duration("duration", 1500*time.Microsecond),
	)

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	out := stripANSI(buf.String())
	if !strings.Contains(out, "INFO") {
		t.Errorf("output missing INFO level: %q", out)
	}
	if !strings.Contains(out, "GET") || !strings.Contains(out, "/health") {
		t.Errorf("output missing method/path: %q", out)
	}
	if !strings.Contains(out, "200") {
		t.Errorf("output missing status: %q", out)
	}
	if !strings.Contains(out, "1.50ms") {
		t.Errorf("output missing duration: %q", out)
	}
}

func TestColoredHandlerAppLog(t *testing.T) {
	var buf bytes.Buffer
	handler := NewColoredHandler(&buf, parseLevel("info"))

	record := slog.NewRecord(
		time.Date(2026, 6, 13, 19, 12, 3, 0, time.UTC),
		slog.LevelInfo,
		"starting nexuslink api",
		0,
	)
	record.AddAttrs(
		slog.String("env", "development"),
		slog.Int("port", 8080),
	)

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	out := stripANSI(buf.String())
	if !strings.Contains(out, "starting nexuslink api") {
		t.Errorf("output missing message: %q", out)
	}
	if !strings.Contains(out, "env=development") {
		t.Errorf("output missing env attr: %q", out)
	}
}

func TestColorsDisabledWithNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	if colorsEnabled() {
		t.Fatal("expected colors disabled when NO_COLOR is set")
	}

	got := paint(red, "error")
	if got != "error" {
		t.Errorf("paint with NO_COLOR = %q, want plain text", got)
	}
}

func stripANSI(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		out.WriteByte(s[i])
	}
	return out.String()
}
