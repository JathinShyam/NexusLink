package shorten

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JathinShyam/NexusLink/internal/storage"
)

type mockRepo struct {
	exists   bool
	mappings map[string]*storage.URLMapping
}

func (m *mockRepo) Create(_ context.Context, mapping *storage.URLMapping) error {
	if m.mappings == nil {
		m.mappings = make(map[string]*storage.URLMapping)
	}
	if _, ok := m.mappings[mapping.ShortCode]; ok {
		return storage.ErrAlreadyExists
	}
	copyMapping := *mapping
	m.mappings[mapping.ShortCode] = &copyMapping
	return nil
}

func (m *mockRepo) GetByShortCode(_ context.Context, shortCode string) (*storage.URLMapping, error) {
	if mapping, ok := m.mappings[shortCode]; ok {
		copyMapping := *mapping
		return &copyMapping, nil
	}
	return nil, storage.ErrNotFound
}

func (m *mockRepo) Exists(_ context.Context, shortCode string) (bool, error) {
	if m.mappings != nil {
		_, ok := m.mappings[shortCode]
		return ok, nil
	}
	return m.exists, nil
}

type mockIDGen struct {
	next uint64
}

func (m *mockIDGen) Next(context.Context) (uint64, error) {
	m.next++
	return m.next, nil
}

func TestServiceCreateGeneratedCode(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockIDGen{}, "http://localhost:8080")

	result, err := svc.Create(context.Background(), Request{
		LongURL: "https://example.com/long/path",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if result.ShortCode == "" {
		t.Fatal("expected generated short code")
	}
	if result.ShortURL != "http://localhost:8080/"+result.ShortCode {
		t.Errorf("ShortURL = %q", result.ShortURL)
	}
	if result.LongURL != "https://example.com/long/path" {
		t.Errorf("LongURL = %q", result.LongURL)
	}
}

func TestServiceCreateCustomAlias(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockIDGen{}, "http://localhost:8080")

	result, err := svc.Create(context.Background(), Request{
		LongURL:     "https://example.com",
		CustomAlias: "my-link",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.ShortCode != "my-link" {
		t.Errorf("ShortCode = %q, want my-link", result.ShortCode)
	}
}

func TestServiceCreateAliasTaken(t *testing.T) {
	repo := &mockRepo{mappings: map[string]*storage.URLMapping{
		"taken": {ShortCode: "taken", LongURL: "https://example.com"},
	}}
	svc := NewService(repo, &mockIDGen{}, "http://localhost:8080")

	_, err := svc.Create(context.Background(), Request{
		LongURL:     "https://example.com",
		CustomAlias: "taken",
	})
	if !errors.Is(err, ErrAliasTaken) {
		t.Fatalf("Create() error = %v, want ErrAliasTaken", err)
	}
}

func TestServiceCreateInvalidURL(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockIDGen{}, "http://localhost:8080")
	_, err := svc.Create(context.Background(), Request{LongURL: "ftp://bad"})
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("Create() error = %v, want ErrInvalidURL", err)
	}
}

func TestServiceCreateWithExpiry(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo, &mockIDGen{}, "http://localhost:8080")
	ttl := int64(3600)

	result, err := svc.Create(context.Background(), Request{
		LongURL:          "https://example.com",
		ExpiresInSeconds: &ttl,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.ExpiresAt == nil {
		t.Fatal("expected expires_at")
	}
	if result.ExpiresAt.Before(time.Now().UTC()) {
		t.Fatal("expires_at should be in the future")
	}
}
