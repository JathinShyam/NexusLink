package shorten

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JathinShyam/NexusLink/internal/encoding/base62"
	"github.com/JathinShyam/NexusLink/internal/idgen"
	"github.com/JathinShyam/NexusLink/internal/storage"
	"github.com/JathinShyam/NexusLink/internal/validation"
)

var (
	ErrInvalidURL   = errors.New("invalid url")
	ErrInvalidAlias = errors.New("invalid custom alias")
	ErrAliasTaken   = errors.New("custom alias already taken")
)

type Request struct {
	LongURL          string
	CustomAlias      string
	ExpiresInSeconds *int64
}

type Response struct {
	ShortCode string
	ShortURL  string
	LongURL   string
	ExpiresAt *time.Time
}

type Service struct {
	repo    storage.Repository
	idgen   idgen.Generator
	baseURL string
}

func NewService(repo storage.Repository, generator idgen.Generator, baseURL string) *Service {
	return &Service{
		repo:    repo,
		idgen:   generator,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (s *Service) Create(ctx context.Context, req Request) (*Response, error) {
	longURL, err := validation.ValidateURL(req.LongURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	var shortCode string
	var snowflakeID *int64

	if req.CustomAlias != "" {
		if err := validation.ValidateShortCode(req.CustomAlias); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidAlias, err)
		}
		exists, err := s.repo.Exists(ctx, req.CustomAlias)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrAliasTaken
		}
		shortCode = req.CustomAlias
	} else {
		id, err := s.idgen.Next(ctx)
		if err != nil {
			return nil, err
		}
		shortCode = base62.Encode(id)
		idCopy := int64(id)
		snowflakeID = &idCopy
	}

	var expiresAt *time.Time
	if req.ExpiresInSeconds != nil {
		if *req.ExpiresInSeconds <= 0 {
			return nil, fmt.Errorf("%w: expires_in_seconds must be positive", ErrInvalidURL)
		}
		t := time.Now().UTC().Add(time.Duration(*req.ExpiresInSeconds) * time.Second)
		expiresAt = &t
	}

	mapping := &storage.URLMapping{
		ShortCode:   shortCode,
		LongURL:     longURL,
		SnowflakeID: snowflakeID,
		ExpiresAt:   expiresAt,
		IsActive:    true,
		HTTPStatus:  httpStatusFound,
	}

	if err := s.repo.Create(ctx, mapping); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, ErrAliasTaken
		}
		return nil, err
	}

	return &Response{
		ShortCode: shortCode,
		ShortURL:  fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		LongURL:   longURL,
		ExpiresAt: expiresAt,
	}, nil
}

const httpStatusFound = 302
