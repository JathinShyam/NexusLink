package redirect

import (
	"errors"
	"net/http"
	"strings"

	"github.com/JathinShyam/NexusLink/internal/httpx"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

type Handler struct {
	resolver *Resolver
}

func NewHandler(resolver *Resolver) *Handler {
	return &Handler{resolver: resolver}
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

	http.Redirect(w, r, target.longURL, target.httpStatus)
}
