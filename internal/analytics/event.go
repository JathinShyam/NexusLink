package analytics

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/JathinShyam/NexusLink/internal/httpx"
)

// Event is a single click on a short link, captured for later analysis.
// It is intentionally small: geo/device enrichment happens downstream in the
// consumer so the redirect hot path stays cheap.
type Event struct {
	ShortCode string    `json:"short_code"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Referrer  string    `json:"referrer,omitempty"`
	RequestID string    `json:"request_id"`
}

// NewEvent captures a click from an inbound redirect request.
func NewEvent(r *http.Request, shortCode string) Event {
	return Event{
		ShortCode: shortCode,
		Timestamp: time.Now().UTC(),
		IP:        httpx.ClientIP(r),
		UserAgent: r.UserAgent(),
		Referrer:  r.Referer(),
		RequestID: requestID(r),
	}
}

// requestID returns the inbound correlation id, generating one when absent so
// the consumer can deduplicate at-least-once deliveries.
func requestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(buf[:])
}
