package accounts

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JathinShyam/NexusLink/internal/auth"
	"github.com/JathinShyam/NexusLink/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail       = errors.New("enter a valid email address")
	ErrWeakPassword       = errors.New("password must be 8-72 characters")
	ErrEmailTaken         = errors.New("an account with that email already exists")
	ErrInvalidCredentials = errors.New("email or password is incorrect")
	ErrInvalidSession     = auth.ErrSessionInvalid
)

const (
	minPasswordBytes = 8
	maxPasswordBytes = 72 // bcrypt only uses the first 72 bytes
	maxEmailBytes    = 254
	maxNameRunes     = 80
	bcryptCost       = 12
)

// dummyHash keeps failed logins for unknown emails roughly as expensive as
// failed logins for known emails.
const dummyHash = "$2a$12$kyiuH3FnYix1jonWRZTYhekm0KMHJVZMsi3xPs5AmU/eEPmp8EMfG"

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuthResult struct {
	Token     string
	ExpiresAt time.Time
	User      User
}

type Service struct {
	repo storage.AccountRepository
	jwt  *auth.JWT
}

func NewService(repo storage.AccountRepository, jwtService *auth.JWT) *Service {
	return &Service{repo: repo, jwt: jwtService}
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (*AuthResult, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	name := normalizeName(displayName, normalizedEmail)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	userID, err := newIdentifier("usr_")
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	user := &storage.User{
		ID:           userID,
		Email:        normalizedEmail,
		DisplayName:  name,
		PasswordHash: string(hash),
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, storage.ErrEmailTaken) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	return s.issueSession(ctx, user)
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if len(password) > maxPasswordBytes {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password[:maxPasswordBytes]))
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueSession(ctx, user)
}

// ValidateSession implements auth.SessionValidator.
func (s *Service) ValidateSession(ctx context.Context, userID, sessionID string) error {
	if userID == "" || sessionID == "" {
		return ErrInvalidSession
	}

	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			return ErrInvalidSession
		}
		return err
	}

	if session.UserID != userID || session.RevokedAt != nil || !session.ExpiresAt.After(time.Now().UTC()) {
		return ErrInvalidSession
	}
	return nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrInvalidSession
	}
	if err := s.repo.RevokeSession(ctx, sessionID); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Service) UserByID(ctx context.Context, userID string) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}
	public := toPublicUser(user)
	return &public, nil
}

func (s *Service) issueSession(ctx context.Context, user *storage.User) (*AuthResult, error) {
	sessionID, err := newIdentifier("ses_")
	if err != nil {
		return nil, fmt.Errorf("generate session id: %w", err)
	}

	expiresAt := time.Now().UTC().Add(s.jwt.TTL())
	session := &storage.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	token, err := s.jwt.IssueSession(user.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &AuthResult{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      toPublicUser(user),
	}, nil
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > maxEmailBytes {
		return "", ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Name != "" || parsed.Address != email {
		return "", ErrInvalidEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || !strings.Contains(parts[1], ".") {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordBytes || len(password) > maxPasswordBytes || strings.TrimSpace(password) == "" {
		return ErrWeakPassword
	}
	return nil
}

func normalizeName(raw, email string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		name = strings.SplitN(email, "@", 2)[0]
	}
	if utf8.RuneCountInString(name) > maxNameRunes {
		runes := []rune(name)
		name = string(runes[:maxNameRunes])
	}
	return name
}

func toPublicUser(user *storage.User) User {
	return User{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
	}
}

func newIdentifier(prefix string) (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(buf[:]), nil
}
