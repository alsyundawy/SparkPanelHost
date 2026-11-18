package auth_test

import (
	"testing"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/auth"
)

func TestArgon2idPasswordHashing(t *testing.T) {
	password := "AetherAdmin@2026!"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !auth.VerifyPassword(password, hash) {
		t.Fatalf("Password verification failed for correct password")
	}

	if auth.VerifyPassword("WrongPassword123", hash) {
		t.Fatalf("Password verification succeeded for incorrect password")
	}
}

func TestSecureTokenAndHash(t *testing.T) {
	token, err := auth.GenerateSecureToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if len(token) != 64 { // 32 bytes hex encoded = 64 chars
		t.Fatalf("Expected token length 64, got %d", len(token))
	}

	hash1 := auth.HashToken(token)
	hash2 := auth.HashToken(token)

	if hash1 != hash2 {
		t.Fatalf("Token hashing is not deterministic")
	}

	if len(hash1) != 64 {
		t.Fatalf("Expected SHA-256 hash length 64, got %d", len(hash1))
	}
}
