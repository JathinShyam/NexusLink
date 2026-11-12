package accounts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JathinShyam/NexusLink/internal/auth"
	"github.com/JathinShyam/NexusLink/internal/storage"
)

type fakeRepo struct {
	users    map[string]*storage.User
	byEmail  map[string]*storage.User
	sessions map[string]*storage.Session
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		users:    make(map[string]*storage.User),
		byEmail:  make(map[string]*storage.User),
		sessions: make(map[string]*storage.Session),
	}
}

func (f *fakeRepo) CreateUser(_ context.Context, user *storage.User) error {
	if _, exists := f.byEmail[user.Email]; exists {
		return storage.ErrEmailTaken
	}
	copyUser := *user
	if copyUser.CreatedAt.IsZero() {
		copyUser.CreatedAt = time.Now().UTC()
	}
	user.CreatedAt = copyUser.CreatedAt
	f.users[user.ID] = &copyUser
	f.byEmail[user.Email] = &copyUser
	return nil
}

func (f *fakeRepo) GetUserByEmail(_ context.Context, email string) (*storage.User, error) {
	if user, ok := f.byEmail[email]; ok {
		copyUser := *user
		return &copyUser, nil
	}
	return nil, storage.ErrNotFound
}

func (f *fakeRepo) GetUserByID(_ context.Context, id string) (*storage.User, error) {
	if user, ok := f.users[id]; ok {
		copyUser := *user
		return &copyUser, nil
	}
	return nil, storage.ErrNotFound
}

func (f *fakeRepo) CreateSession(_ context.Context, session *storage.Session) error {
	copySession := *session
	if copySession.CreatedAt.IsZero() {
		copySession.CreatedAt = time.Now().UTC()
	}
	session.CreatedAt = copySession.CreatedAt
	f.sessions[session.ID] = &copySession
	return nil
}

func (f *fakeRepo) GetSession(_ context.Context, id string) (*storage.Session, error) {
	if session, ok := f.sessions[id]; ok {
		copySession := *session
		return &copySession, nil
	}
	return nil, storage.ErrSessionNotFound
}

func (f *fakeRepo) RevokeSession(_ context.Context, id string) error {
	session, ok := f.sessions[id]
	if !ok || session.RevokedAt != nil {
		return storage.ErrSessionNotFound
	}
	now := time.Now().UTC()
	session.RevokedAt = &now
	return nil
}

func TestRegisterAndLogin(t *testing.T) {
	repo := newFakeRepo()
	jwtService := auth.NewJWT("test-secret", time.Hour)
	svc := NewService(repo, jwtService)
	ctx := context.Background()

	result, err := svc.Register(ctx, "  Alice@Example.COM ", "correct horse", "")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.User.Email != "alice@example.com" {
		t.Errorf("Email = %q, want alice@example.com", result.User.Email)
	}
	if result.User.DisplayName != "alice" {
		t.Errorf("DisplayName = %q, want alice", result.User.DisplayName)
	}
	if result.Token == "" {
		t.Fatal("expected a token")
	}
	if result.ExpiresAt.Before(time.Now()) {
		t.Error("expected a future expiry")
	}

	claims, err := jwtService.Parse(result.Token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != result.User.ID {
		t.Errorf("UserID = %q, want %q", claims.UserID, result.User.ID)
	}
	if claims.SessionID == "" {
		t.Fatal("expected a session ID claim")
	}
	if err := svc.ValidateSession(ctx, result.User.ID, claims.SessionID); err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}

	login, err := svc.Login(ctx, "ALICE@example.com", "correct horse")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if login.User.ID != result.User.ID {
		t.Errorf("Login user ID = %q, want %q", login.User.ID, result.User.ID)
	}
	if login.Token == result.Token {
		t.Error("expected a new session token")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, auth.NewJWT("test-secret", time.Hour))
	ctx := context.Background()

	if _, err := svc.Register(ctx, "alice@example.com", "password-1", ""); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if _, err := svc.Register(ctx, "ALICE@example.com", "password-2", ""); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("second Register() error = %v, want ErrEmailTaken", err)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, auth.NewJWT("test-secret", time.Hour))
	ctx := context.Background()

	if _, err := svc.Register(ctx, "alice@example.com", "correct horse", ""); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if _, err := svc.Login(ctx, "alice@example.com", "wrong horse"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := svc.Login(ctx, "nobody@example.com", "correct horse"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() unknown email error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	repo := newFakeRepo()
	jwtService := auth.NewJWT("test-secret", time.Hour)
	svc := NewService(repo, jwtService)
	ctx := context.Background()

	result, err := svc.Register(ctx, "alice@example.com", "correct horse", "")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	claims, err := jwtService.Parse(result.Token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if err := svc.Logout(ctx, claims.SessionID); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if err := svc.ValidateSession(ctx, result.User.ID, claims.SessionID); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("ValidateSession() error = %v, want ErrInvalidSession", err)
	}
}

func TestValidateSessionExpired(t *testing.T) {
	repo := newFakeRepo()
	jwtService := auth.NewJWT("test-secret", time.Hour)
	svc := NewService(repo, jwtService)
	ctx := context.Background()

	result, err := svc.Register(ctx, "alice@example.com", "correct horse", "")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	claims, err := jwtService.Parse(result.Token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	repo.sessions[claims.SessionID].ExpiresAt = time.Now().Add(-time.Minute)

	if err := svc.ValidateSession(ctx, result.User.ID, claims.SessionID); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("ValidateSession() error = %v, want ErrInvalidSession", err)
	}
}

func TestRegisterValidation(t *testing.T) {
	svc := NewService(newFakeRepo(), auth.NewJWT("test-secret", time.Hour))
	ctx := context.Background()

	cases := []struct {
		name     string
		email    string
		password string
		want     error
	}{
		{name: "empty email", email: "", password: "password-1", want: ErrInvalidEmail},
		{name: "invalid email", email: "not-an-email", password: "password-1", want: ErrInvalidEmail},
		{name: "short password", email: "alice@example.com", password: "short", want: ErrWeakPassword},
		{name: "long password", email: "alice@example.com", password: string(make([]byte, 73)), want: ErrWeakPassword},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.Register(ctx, tc.email, tc.password, ""); !errors.Is(err, tc.want) {
				t.Fatalf("Register() error = %v, want %v", err, tc.want)
			}
		})
	}
}
