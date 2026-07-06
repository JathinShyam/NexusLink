package links

import (
	"errors"
	"net/http"

	"github.com/JathinShyam/NexusLink/internal/auth"
	"github.com/JathinShyam/NexusLink/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "short_code is required")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	if err := h.service.Delete(r.Context(), shortCode, userID); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "link not found")
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrNotOwned):
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete link")
	}
}
