package links

import (
	"context"
	"errors"
	"fmt"

	"github.com/JathinShyam/NexusLink/internal/cache"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

var (
	ErrNotFound  = errors.New("link not found")
	ErrForbidden = errors.New("not allowed to delete this link")
	ErrNotOwned  = errors.New("anonymous links cannot be deleted")
)

type Service struct {
	repo  storage.Repository
	cache cache.URLCache
}

func NewService(repo storage.Repository, urlCache cache.URLCache) *Service {
	return &Service{repo: repo, cache: urlCache}
}

func (s *Service) Delete(ctx context.Context, shortCode, userID string) error {
	mapping, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	if mapping.UserID == nil {
		return ErrNotOwned
	}
	if *mapping.UserID != userID {
		return ErrForbidden
	}

	if err := s.repo.Deactivate(ctx, shortCode); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("deactivate link: %w", err)
	}

	_ = s.cache.Delete(ctx, shortCode)
	return nil
}
