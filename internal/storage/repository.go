package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("url mapping not found")
	ErrAlreadyExists = errors.New("short code already exists")
)

type URLMapping struct {
	ShortCode   string
	LongURL     string
	SnowflakeID *int64
	UserID      *string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	IsActive    bool
	HTTPStatus  int
}

type Repository interface {
	Create(ctx context.Context, mapping *URLMapping) error
	GetByShortCode(ctx context.Context, shortCode string) (*URLMapping, error)
	Exists(ctx context.Context, shortCode string) (bool, error)
}
