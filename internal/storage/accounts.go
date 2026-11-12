package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrEmailTaken      = errors.New("email already registered")
	ErrSessionNotFound = errors.New("session not found")
)

type User struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// AccountRepository stores user accounts and revocable login sessions.
// It is kept separate from Repository so link storage mocks stay small.
type AccountRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	RevokeSession(ctx context.Context, id string) error
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user *User) error {
	const query = `
		INSERT INTO users (id, email, display_name, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.DisplayName,
		user.PasswordHash,
	).Scan(&user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, email, display_name, password_hash, created_at
		FROM users
		WHERE email = $1`

	var user User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	const query = `
		SELECT id, email, display_name, password_hash, created_at
		FROM users
		WHERE id = $1`

	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session *Session) error {
	const query = `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		session.ID,
		session.UserID,
		session.ExpiresAt,
	).Scan(&session.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetSession(ctx context.Context, id string) (*Session, error) {
	const query = `
		SELECT id, user_id, created_at, expires_at, revoked_at
		FROM sessions
		WHERE id = $1`

	var session Session
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &session, nil
}

func (r *PostgresRepository) RevokeSession(ctx context.Context, id string) error {
	const query = `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}
