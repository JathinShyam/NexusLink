package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, mapping *URLMapping) error {
	const query = `
		INSERT INTO url_mappings (
			short_code, long_url, snowflake_id, user_id, expires_at, is_active, http_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		mapping.ShortCode,
		mapping.LongURL,
		mapping.SnowflakeID,
		mapping.UserID,
		mapping.ExpiresAt,
		mapping.IsActive,
		mapping.HTTPStatus,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("insert url mapping: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByShortCode(ctx context.Context, shortCode string) (*URLMapping, error) {
	const query = `
		SELECT short_code, long_url, snowflake_id, user_id, created_at, expires_at, is_active, http_status
		FROM url_mappings
		WHERE short_code = $1`

	var mapping URLMapping
	err := r.pool.QueryRow(ctx, query, shortCode).Scan(
		&mapping.ShortCode,
		&mapping.LongURL,
		&mapping.SnowflakeID,
		&mapping.UserID,
		&mapping.CreatedAt,
		&mapping.ExpiresAt,
		&mapping.IsActive,
		&mapping.HTTPStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get url mapping: %w", err)
	}
	return &mapping, nil
}

func (r *PostgresRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM url_mappings WHERE short_code = $1)`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, shortCode).Scan(&exists); err != nil {
		return false, fmt.Errorf("check short code exists: %w", err)
	}
	return exists, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
