package shorten

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JathinShyam/NexusLink/internal/storage"
)

func TestHandlerCreateSuccess(t *testing.T) {
	repo := &mockRepo{}
	handler := NewHandler(NewService(repo, &mockIDGen{}, "http://localhost:8080"))

	body := `{"long_url":"https://example.com/path"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp shortenResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ShortCode == "" || resp.ShortURL == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestHandlerCreateInvalidJSON(t *testing.T) {
	handler := NewHandler(NewService(&mockRepo{}, &mockIDGen{}, "http://localhost:8080"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()
	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandlerCreateConflict(t *testing.T) {
	repo := &mockRepo{mappings: map[string]*storage.URLMapping{
		"dup": {ShortCode: "dup", LongURL: "https://example.com"},
	}}
	handler := NewHandler(NewService(repo, &mockIDGen{}, "http://localhost:8080"))

	body := `{"long_url":"https://example.com","custom_alias":"dup"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	handler.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
