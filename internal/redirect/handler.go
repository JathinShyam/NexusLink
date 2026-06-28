package redirect

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/JathinShyam/NexusLink/internal/httpx"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

type Handler struct {
	repo storage.Repository
}

func NewHandler(repo storage.Repository) *Handler {
	return &Handler{repo: repo}
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

	mapping, err := h.repo.GetByShortCode(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "short link not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve short link")
		return
	}

	if !mapping.IsActive {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "short link not found")
		return
	}

	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now().UTC()) {
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "short link has expired")
		return
	}

	status := mapping.HTTPStatus
	if status == 0 {
		status = http.StatusFound
	}

	http.Redirect(w, r, mapping.LongURL, status)
}
