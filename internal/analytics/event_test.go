package analytics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewEventFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 10.0.0.1")
	req.Header.Set("User-Agent", "test-agent/1.0")
	req.Header.Set("Referer", "https://example.com/source")
	req.Header.Set("X-Request-Id", "req-123")

	event := NewEvent(req, "abc123")

	if event.ShortCode != "abc123" {
		t.Errorf("ShortCode = %q, want abc123", event.ShortCode)
	}
	if event.IP != "203.0.113.7" {
		t.Errorf("IP = %q, want first X-Forwarded-For hop", event.IP)
	}
	if event.UserAgent != "test-agent/1.0" {
		t.Errorf("UserAgent = %q", event.UserAgent)
	}
	if event.Referrer != "https://example.com/source" {
		t.Errorf("Referrer = %q", event.Referrer)
	}
	if event.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want inbound header value", event.RequestID)
	}
	if event.Timestamp.IsZero() {
		t.Error("Timestamp not set")
	}
}

func TestNewEventGeneratesRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	req.RemoteAddr = "198.51.100.4:55555"

	event := NewEvent(req, "abc123")

	if event.IP != "198.51.100.4" {
		t.Errorf("IP = %q, want remote host without port", event.IP)
	}
	if len(event.RequestID) != 32 {
		t.Errorf("RequestID = %q, want 32-char hex", event.RequestID)
	}
}
