package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	TOTPPeriod = 30
	TOTPDigits = 6
)

var base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// GenerateTOTPSecret generates a cryptographically secure 20-byte base32 secret.
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes for TOTP: %w", err)
	}
	return base32Encoding.EncodeToString(bytes), nil
}

// GenerateTOTPURI generates an otpauth:// URL for QR code generation in authenticator apps.
func GenerateTOTPURI(username, issuer, secret string) string {
	label := url.PathEscape(issuer + ":" + username)
	v := url.Values{}
	v.Set("secret", secret)
	v.Set("issuer", issuer)
	v.Set("algorithm", "SHA1")
	v.Set("digits", strconv.Itoa(TOTPDigits))
	v.Set("period", strconv.Itoa(TOTPPeriod))
	return fmt.Sprintf("otpauth://totp/%s?%s", label, v.Encode())
}

// ComputeTOTP calculates the current 6-digit TOTP code for a secret at a given timestamp.
func ComputeTOTP(secret string, t time.Time) (string, error) {
	key, err := base32Encoding.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", fmt.Errorf("invalid base32 secret: %w", err)
	}

	unixSec := t.Unix()
	if unixSec < 0 {
		return "", errors.New("invalid timestamp")
	}
	counter := uint64(unixSec / TOTPPeriod) // #nosec G115 - timestamp is verified non-negative
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key) // #nosec G505 - RFC 6238 mandates HMAC-SHA1 algorithm for standard TOTP
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	codeInt := (binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff) % 1000000

	return fmt.Sprintf("%06d", codeInt), nil
}

// VerifyTOTP checks a 6-digit code against a secret with a configurable time-step window.
// window = 1 checks t-1, t, and t+1 (±30 seconds) to account for client clock drift.
func VerifyTOTP(secret, userCode string, window int) bool {
	userCode = strings.TrimSpace(userCode)
	if len(userCode) != TOTPDigits {
		return false
	}

	now := time.Now()
	for i := -window; i <= window; i++ {
		t := now.Add(time.Duration(i*TOTPPeriod) * time.Second)
		expectedCode, err := ComputeTOTP(secret, t)
		if err != nil {
			return false
		}
		if subtle.ConstantTimeCompare([]byte(userCode), []byte(expectedCode)) == 1 {
			return true
		}
	}
	return false
}

// EncryptSecret encrypts a secret with AES-256-GCM using a 32-byte key.
func EncryptSecret(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be exactly 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptSecret decrypts an AES-256-GCM ciphertext produced by EncryptSecret.
func DecryptSecret(ciphertextHex string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be exactly 32 bytes")
	}

	data, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
