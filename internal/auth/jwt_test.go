package auth

import (
	"testing"
	"time"
)

func TestJWTIssueAndParse(t *testing.T) {
	jwtService := NewJWT("test-secret", time.Hour)
	token, err := jwtService.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	claims, err := jwtService.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want user-123", claims.UserID)
	}
}

func TestJWTParseInvalid(t *testing.T) {
	jwtService := NewJWT("test-secret", time.Hour)
	if _, err := jwtService.Parse("not-a-token"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}
