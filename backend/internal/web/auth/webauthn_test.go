package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestWebAuthnChallengeAndVerification(t *testing.T) {
	challenge, err := GenerateWebAuthnChallenge()
	if err != nil {
		t.Fatalf("GenerateWebAuthnChallenge failed: %v", err)
	}
	if len(challenge) == 0 {
		t.Fatal("empty challenge")
	}

	cd := ClientData{
		Type:      "webauthn.create",
		Challenge: challenge,
		Origin:    "https://panel.example.com:2087",
	}
	rawJSON, err := json.Marshal(cd)
	if err != nil {
		t.Fatalf("failed to marshal clientData: %v", err)
	}

	// 1. Success case
	err = VerifyClientData(rawJSON, "webauthn.create", challenge, "https://panel.example.com:2087")
	if err != nil {
		t.Fatalf("expected verification to pass, got: %v", err)
	}

	// 2. Type mismatch
	err = VerifyClientData(rawJSON, "webauthn.get", challenge, "https://panel.example.com:2087")
	if err == nil {
		t.Fatal("expected error for type mismatch, got nil")
	}

	// 3. Challenge mismatch
	err = VerifyClientData(rawJSON, "webauthn.create", "wrong_challenge", "https://panel.example.com:2087")
	if err == nil {
		t.Fatal("expected error for challenge mismatch, got nil")
	}

	// 4. Origin mismatch
	err = VerifyClientData(rawJSON, "webauthn.create", challenge, "https://evil.com")
	if err == nil {
		t.Fatal("expected error for origin mismatch, got nil")
	}

	hash := ClientDataHash(rawJSON)
	if hash == [32]byte{} {
		t.Fatal("expected non-empty hash")
	}
}

func TestBase64DecodingHelper(t *testing.T) {
	data := "test-credential-id"
	encoded := base64.RawURLEncoding.EncodeToString([]byte(data))
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if string(decoded) != data {
		t.Fatalf("decoded (%s) != original (%s)", string(decoded), data)
	}
}
