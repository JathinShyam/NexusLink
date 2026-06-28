package shorten

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/JathinShyam/NexusLink/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type shortenRequest struct {
	LongURL          string `json:"long_url"`
	CustomAlias      string `json:"custom_alias"`
	ExpiresInSeconds *int64 `json:"expires_in_seconds"`
}

type shortenResponse struct {
	ShortCode string     `json:"short_code"`
	ShortURL  string     `json:"short_url"`
	LongURL   string     `json:"long_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
		return
	}

	result, err := h.service.Create(r.Context(), Request{
		LongURL:          body.LongURL,
		CustomAlias:      body.CustomAlias,
		ExpiresInSeconds: body.ExpiresInSeconds,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, shortenResponse{
		ShortCode: result.ShortCode,
		ShortURL:  result.ShortURL,
		LongURL:   result.LongURL,
		ExpiresAt: result.ExpiresAt,
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidURL):
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_URL", err.Error())
	case errors.Is(err, ErrInvalidAlias):
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ALIAS", err.Error())
	case errors.Is(err, ErrAliasTaken):
		httpx.WriteError(w, http.StatusConflict, "ALIAS_TAKEN", "custom alias is already taken")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create short url")
	}
}
