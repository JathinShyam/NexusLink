package links

import (
	"context"
	"errors"
	"testing"

	"github.com/JathinShyam/NexusLink/internal/cache"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

type mockRepo struct {
	mappings map[string]*storage.URLMapping
}

func (m *mockRepo) Create(context.Context, *storage.URLMapping) error { return nil }
func (m *mockRepo) Exists(context.Context, string) (bool, error)      { return false, nil }

func (m *mockRepo) GetByShortCode(_ context.Context, shortCode string) (*storage.URLMapping, error) {
	if mapping, ok := m.mappings[shortCode]; ok {
		copyMapping := *mapping
		return &copyMapping, nil
	}
	return nil, storage.ErrNotFound
}

func (m *mockRepo) Deactivate(_ context.Context, shortCode string) error {
	mapping, ok := m.mappings[shortCode]
	if !ok || !mapping.IsActive {
		return storage.ErrNotFound
	}
	mapping.IsActive = false
	return nil
}

func TestServiceDeleteOwnedLink(t *testing.T) {
	userID := "alice"
	repo := &mockRepo{mappings: map[string]*storage.URLMapping{
		"abc": {
			ShortCode: "abc",
			LongURL:   "https://example.com",
			UserID:    &userID,
			IsActive:  true,
		},
	}}
	svc := NewService(repo, cache.NoopCache{})

	if err := svc.Delete(context.Background(), "abc", "alice"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if repo.mappings["abc"].IsActive {
		t.Fatal("expected link to be deactivated")
	}
}

func TestServiceDeleteForbidden(t *testing.T) {
	owner := "alice"
	repo := &mockRepo{mappings: map[string]*storage.URLMapping{
		"abc": {ShortCode: "abc", UserID: &owner, IsActive: true},
	}}
	svc := NewService(repo, cache.NoopCache{})

	err := svc.Delete(context.Background(), "abc", "bob")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Delete() error = %v, want ErrForbidden", err)
	}
}
