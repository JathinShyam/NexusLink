package redirect

import (
	"errors"
	"net/http"
	"strings"

	"github.com/JathinShyam/NexusLink/internal/analytics"
	"github.com/JathinShyam/NexusLink/internal/httpx"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

type Handler struct {
	resolver  *Resolver
	publisher analytics.Publisher
}

// NewHandler returns a redirect handler. A nil publisher is treated as a no-op
// so analytics stays optional.
func NewHandler(resolver *Resolver, publisher analytics.Publisher) *Handler {
	if publisher == nil {
		publisher = analytics.NoopPublisher{}
	}
	return &Handler{resolver: resolver, publisher: publisher}
}

func (h *Handler) Serve(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		shortCode = strings.TrimPrefix(r.URL.Path, "/")
	}
	if shortCode == "" || strings.Contains(shortCode, "/") {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "short link not found")
		return
	}

	target, err := h.resolver.Resolve(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "short link not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve short link")
		return
	}

	// Non-blocking: analytics must never affect redirect latency (NF3).
	h.publisher.Publish(r.Context(), analytics.NewEvent(r, shortCode))

	http.Redirect(w, r, target.longURL, target.httpStatus)
}
