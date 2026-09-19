package auth_test

import (
	"testing"
	"time"

	"github.com/vignesh/livepoll/internal/auth"
)

func TestPasswordHashing(t *testing.T) {
	plain := "supersecret123"
	hash, err := auth.HashPassword(plain)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == plain {
		t.Fatal("hashed password must not match plain text")
	}

	if !auth.CheckPassword(hash, plain) {
		t.Fatal("password verification failed for matching password")
	}

	if auth.CheckPassword(hash, "wrongpassword") {
		t.Fatal("password verification succeeded for incorrect password")
	}
}

func TestJWTTokenLifecycle(t *testing.T) {
	secret := []byte("a-very-long-secret-key-at-least-32-bytes-long!")
	tm := auth.NewTokenManager(secret, 1*time.Hour)

	userID := "65f0a1b2c3d4e5f6a7b8c9d0"
	email := "test@example.com"
	name := "Tester"

	tokenStr, expiresAt, err := tm.Issue(userID, email, name)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("issued token string must not be empty")
	}

	if expiresAt.Before(time.Now()) {
		t.Fatal("token expiry must be in the future")
	}

	// Parse and verify claims
	claims, err := tm.Parse(tokenStr)
	if err != nil {
		t.Fatalf("failed to parse valid token: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("expected subject %s, got %s", userID, claims.Subject)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.Name != name {
		t.Errorf("expected name %s, got %s", name, claims.Name)
	}

	// Verify tampering rejection
	tampered := tokenStr + "tampered"
	if _, err := tm.Parse(tampered); err == nil {
		t.Fatal("expected error parsing tampered token, got nil")
	}

	// Verify wrong secret rejection
	wrongTM := auth.NewTokenManager([]byte("completely-different-signing-secret!"), 1*time.Hour)
	if _, err := wrongTM.Parse(tokenStr); err == nil {
		t.Fatal("expected error parsing token with different secret, got nil")
	}

	// Verify expired token rejection
	expiredTM := auth.NewTokenManager(secret, -1*time.Minute)
	expiredToken, _, err := expiredTM.Issue(userID, email, name)
	if err != nil {
		t.Fatalf("failed to issue expired token: %v", err)
	}
	if _, err := tm.Parse(expiredToken); err == nil {
		t.Fatal("expected error parsing expired token, got nil")
	}
}
