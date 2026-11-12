package accounts

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/JathinShyam/NexusLink/internal/auth"
	"github.com/JathinShyam/NexusLink/internal/httpx"
)

const maxBodyBytes = 16 << 10

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

type meResponse struct {
	User User `json:"user"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if !decodeJSON(w, r, &body) {
		return
	}

	result, err := h.service.Register(r.Context(), body.Email, body.Password, body.DisplayName)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, authResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		User:      result.User,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if !decodeJSON(w, r, &body) {
		return
	}

	result, err := h.service.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, authResponse{
		Token:     result.Token,
		ExpiresAt: result.ExpiresAt,
		User:      result.User,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := auth.SessionIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "sign in required")
		return
	}

	if err := h.service.Logout(r.Context(), sessionID); err != nil {
		if errors.Is(err, ErrInvalidSession) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to sign out")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "sign in required")
		return
	}

	user, err := h.service.UserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrInvalidSession) {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "account not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load account")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, meResponse{User: *user})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dst); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body must be valid JSON")
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmail):
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_EMAIL", ErrInvalidEmail.Error())
	case errors.Is(err, ErrWeakPassword):
		httpx.WriteError(w, http.StatusBadRequest, "WEAK_PASSWORD", ErrWeakPassword.Error())
	case errors.Is(err, ErrEmailTaken):
		httpx.WriteError(w, http.StatusConflict, "EMAIL_TAKEN", ErrEmailTaken.Error())
	case errors.Is(err, ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", ErrInvalidCredentials.Error())
	case errors.Is(err, ErrInvalidSession):
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "session expired; sign in again")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
	}
}
