package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type ClientData struct {
	Type        string `json:"type"`
	Challenge   string `json:"challenge"`
	Origin      string `json:"origin"`
	CrossOrigin bool   `json:"crossOrigin,omitempty"`
}

type PasskeyRegistrationRequest struct {
	DeviceName     string `json:"device_name"`
	CredentialID   string `json:"credential_id"` // base64url encoded
	PublicKey      string `json:"public_key"`    // base64url encoded
	ClientDataJSON string `json:"client_data_json"`
}

type PasskeyLoginRequest struct {
	CredentialID      string `json:"credential_id"`
	ClientDataJSON    string `json:"client_data_json"`
	AuthenticatorData string `json:"authenticator_data"`
	Signature         string `json:"signature"`
}

// GenerateWebAuthnChallenge creates a 32-byte base64url-encoded challenge.
func GenerateWebAuthnChallenge() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", fmt.Errorf("failed to generate random challenge: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// VerifyClientData validates the clientDataJSON against expected type, challenge, and origin.
func VerifyClientData(clientDataJSONRaw []byte, expectedType, expectedChallenge, expectedOrigin string) error {
	var cd ClientData
	if err := json.Unmarshal(clientDataJSONRaw, &cd); err != nil {
		return fmt.Errorf("malformed clientDataJSON: %w", err)
	}

	if cd.Type != expectedType {
		return fmt.Errorf("invalid client data type: expected %q, got %q", expectedType, cd.Type)
	}

	if subtle.ConstantTimeCompare([]byte(cd.Challenge), []byte(expectedChallenge)) != 1 {
		return errors.New("challenge mismatch in clientDataJSON")
	}

	// In production, verify exact origin or allowed origin
	if expectedOrigin != "" && !strings.EqualFold(cd.Origin, expectedOrigin) {
		return fmt.Errorf("origin mismatch: expected %q, got %q", expectedOrigin, cd.Origin)
	}

	return nil
}

// ClientDataHash calculates SHA-256 of clientDataJSON.
func ClientDataHash(clientDataJSON []byte) [32]byte {
	return sha256.Sum256(clientDataJSON)
}
