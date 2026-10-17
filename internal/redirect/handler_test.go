package redirect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JathinShyam/NexusLink/internal/analytics"
	"github.com/JathinShyam/NexusLink/internal/cache"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

type recordingPublisher struct {
	events []analytics.Event
}

func (p *recordingPublisher) Publish(_ context.Context, e analytics.Event) {
	p.events = append(p.events, e)
}

type mockRepo struct {
	mapping *storage.URLMapping
	err     error
}

func (m *mockRepo) Create(context.Context, *storage.URLMapping) error { return nil }
func (m *mockRepo) Exists(context.Context, string) (bool, error)      { return false, nil }

func (m *mockRepo) Deactivate(context.Context, string) error { return nil }

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
	resolver := NewResolver(&mockRepo{
		mapping: &storage.URLMapping{
			ShortCode:  "abc",
			LongURL:    "https://example.com",
			IsActive:   true,
			HTTPStatus: http.StatusFound,
		},
	}, cache.NoopCache{}, nil)
	handler := NewHandler(resolver, nil)

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
	handler := NewHandler(NewResolver(&mockRepo{}, cache.NoopCache{}, nil), nil)
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	handler.Serve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerRedirectExpired(t *testing.T) {
	expired := time.Now().UTC().Add(-time.Hour)
	resolver := NewResolver(&mockRepo{
		mapping: &storage.URLMapping{
			ShortCode: "expired",
			LongURL:   "https://example.com",
			IsActive:  true,
			ExpiresAt: &expired,
		},
	}, cache.NoopCache{}, nil)
	handler := NewHandler(resolver, nil)

	req := httptest.NewRequest(http.MethodGet, "/expired", nil)
	rec := httptest.NewRecorder()
	handler.Serve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerPublishesClickEventOnSuccess(t *testing.T) {
	publisher := &recordingPublisher{}
	resolver := NewResolver(&mockRepo{
		mapping: &storage.URLMapping{
			ShortCode:  "abc",
			LongURL:    "https://example.com",
			IsActive:   true,
			HTTPStatus: http.StatusFound,
		},
	}, cache.NoopCache{}, nil)
	handler := NewHandler(resolver, publisher)

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	handler.Serve(httptest.NewRecorder(), req)

	if len(publisher.events) != 1 {
		t.Fatalf("published %d events, want 1", len(publisher.events))
	}
	if got := publisher.events[0].ShortCode; got != "abc" {
		t.Errorf("event ShortCode = %q, want abc", got)
	}
}

func TestHandlerDoesNotPublishOnMiss(t *testing.T) {
	publisher := &recordingPublisher{}
	handler := NewHandler(NewResolver(&mockRepo{}, cache.NoopCache{}, nil), publisher)

	handler.Serve(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))

	if len(publisher.events) != 0 {
		t.Fatalf("published %d events on miss, want 0", len(publisher.events))
	}
}

var _ storage.Repository = (*mockRepo)(nil)
