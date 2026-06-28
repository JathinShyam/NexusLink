package redirect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JathinShyam/NexusLink/internal/storage"
)

type mockRepo struct {
	mapping *storage.URLMapping
	err     error
}

func (m *mockRepo) Create(context.Context, *storage.URLMapping) error { return nil }
func (m *mockRepo) Exists(context.Context, string) (bool, error)      { return false, nil }

func (m *mockRepo) GetByShortCode(context.Context, string) (*storage.URLMapping, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.mapping == nil {
		return nil, storage.ErrNotFound
	}
	copyMapping := *m.mapping
	return &copyMapping, nil
}

func TestHandlerRedirectSuccess(t *testing.T) {
	handler := NewHandler(&mockRepo{
		mapping: &storage.URLMapping{
			ShortCode:  "abc",
			LongURL:    "https://example.com",
			IsActive:   true,
			HTTPStatus: http.StatusFound,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rec := httptest.NewRecorder()
	handler.Serve(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if got := rec.Header().Get("Location"); got != "https://example.com" {
		t.Errorf("Location = %q", got)
	}
}

func TestHandlerRedirectNotFound(t *testing.T) {
	handler := NewHandler(&mockRepo{})
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	handler.Serve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerRedirectExpired(t *testing.T) {
	expired := time.Now().UTC().Add(-time.Hour)
	handler := NewHandler(&mockRepo{
		mapping: &storage.URLMapping{
			ShortCode: "expired",
			LongURL:   "https://example.com",
			IsActive:  true,
			ExpiresAt: &expired,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/expired", nil)
	rec := httptest.NewRecorder()
	handler.Serve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

var _ storage.Repository = (*mockRepo)(nil)
