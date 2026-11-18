package auth

import (
	"strings"
	"testing"
	"time"
)

func TestTOTPGenerationAndVerification(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}
	if len(secret) == 0 {
		t.Fatal("secret is empty")
	}

	uri := GenerateTOTPURI("admin", "AetherPanel", secret)
	if !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("unexpected TOTP URI: %s", uri)
	}
	if !strings.Contains(uri, "secret="+secret) {
		t.Fatalf("URI missing secret parameter: %s", uri)
	}

	// Compute code at current time
	code, err := ComputeTOTP(secret, time.Now())
	if err != nil {
		t.Fatalf("ComputeTOTP failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6-digit code, got: %s", code)
	}

	// Verify valid code
	if !VerifyTOTP(secret, code, 1) {
		t.Fatal("VerifyTOTP returned false for valid code")
	}

	// Verify invalid code
	if VerifyTOTP(secret, "000000", 0) && code != "000000" {
		t.Fatal("VerifyTOTP returned true for bogus code")
	}
}

func TestTOTPSecretEncryption(t *testing.T) {
	key := []byte("01234567890123456789012345678901") // 32 bytes
	plaintext := "JBSWY3DPEHPK3PXP"

	ciphertext, err := EncryptSecret(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptSecret failed: %v", err)
	}

	decrypted, err := DecryptSecret(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptSecret failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("decrypted (%s) != plaintext (%s)", decrypted, plaintext)
	}

	// Test with invalid key
	wrongKey := []byte("99999999999999999999999999999999")
	_, err = DecryptSecret(ciphertext, wrongKey)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key, got nil")
	}
}
