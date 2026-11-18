package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/auth"
	"github.com/alsyundawy/SparkPanelHost/backend/internal/web/store"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var defaultPanelKey = []byte("aetherpanel-secret-key-32bytes!")

func (s *Server) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	u := GetUserFromContext(r.Context())
	if u == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate secret")
		return
	}

	uri := auth.GenerateTOTPURI(u.Username, "AetherPanel", secret)
	writeJSON(w, http.StatusOK, map[string]string{
		"secret":      secret,
		"otpauth_url": uri,
	})
}

func (s *Server) EnableTOTP(w http.ResponseWriter, r *http.Request) {
	u := GetUserFromContext(r.Context())
	if u == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if !auth.VerifyTOTP(req.Secret, req.Code, 1) {
		writeJSONError(w, http.StatusBadRequest, "kode TOTP salah atau kadaluarsa")
		return
	}

	encSecret, err := auth.EncryptSecret(req.Secret, defaultPanelKey)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to encrypt secret")
		return
	}

	if err := s.Store.UpdateUserTOTP(r.Context(), u.ID, encSecret, true); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	_ = s.Store.AppendAuditLog(r.Context(), u.ID, nil, "2fa.totp_enabled", "user:"+u.ID, r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]string{"message": "TOTP 2FA berhasil diaktifkan"})
}

func (s *Server) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	u := GetUserFromContext(r.Context())
	if u == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := s.Store.UpdateUserTOTP(r.Context(), u.ID, "", false); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	_ = s.Store.AppendAuditLog(r.Context(), u.ID, nil, "2fa.totp_disabled", "user:"+u.ID, r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]string{"message": "TOTP 2FA berhasil dinonaktifkan"})
}

func (s *Server) RegisterPasskeyChallenge(w http.ResponseWriter, r *http.Request) {
	challenge, err := auth.GenerateWebAuthnChallenge()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate challenge")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"challenge": challenge})
}

func (s *Server) RegisterPasskeyVerify(w http.ResponseWriter, r *http.Request) {
	u := GetUserFromContext(r.Context())
	if u == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		DeviceName        string `json:"device_name"`
		CredentialID      string `json:"credential_id"`
		PublicKey         string `json:"public_key"`
		ClientDataJSON    string `json:"client_data_json"`
		ExpectedChallenge string `json:"expected_challenge"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request")
		return
	}

	clientDataBytes, err := base64.RawURLEncoding.DecodeString(req.ClientDataJSON)
	if err != nil {
		clientDataBytes, _ = base64.StdEncoding.DecodeString(req.ClientDataJSON)
	}
	if len(clientDataBytes) > 0 {
		if err := auth.VerifyClientData(clientDataBytes, "webauthn.create", req.ExpectedChallenge, ""); err != nil {
			writeJSONError(w, http.StatusBadRequest, "verifikasi WebAuthn gagal: "+err.Error())
			return
		}
	}

	credIDBytes, err := base64.RawURLEncoding.DecodeString(req.CredentialID)
	if err != nil {
		credIDBytes = []byte(req.CredentialID)
	}
	pubKeyBytes, err := base64.RawURLEncoding.DecodeString(req.PublicKey)
	if err != nil {
		pubKeyBytes = []byte(req.PublicKey)
	}

	passkey := &store.UserPasskey{
		ID:           uuid.New().String(),
		UserID:       u.ID,
		CredentialID: credIDBytes,
		PublicKey:    pubKeyBytes,
		SignCount:    0,
		DeviceName:   req.DeviceName,
	}

	if err := s.Store.CreateUserPasskey(r.Context(), passkey); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "gagal menyimpan passkey: "+err.Error())
		return
	}

	_ = s.Store.AppendAuditLog(r.Context(), u.ID, nil, "passkey.created", "passkey:"+passkey.ID, r.RemoteAddr)
	writeJSON(w, http.StatusCreated, passkey)
}

func (s *Server) ListPasskeys(w http.ResponseWriter, r *http.Request) {
	u := GetUserFromContext(r.Context())
	if u == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	passkeys, err := s.Store.GetUserPasskeys(r.Context(), u.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, passkeys)
}

func (s *Server) DeletePasskey(w http.ResponseWriter, r *http.Request) {
	u := GetUserFromContext(r.Context())
	if u == nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	passkeyID := chi.URLParam(r, "id")
	if err := s.Store.DeleteUserPasskey(r.Context(), passkeyID, u.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	_ = s.Store.AppendAuditLog(r.Context(), u.ID, nil, "passkey.deleted", "passkey:"+passkeyID, r.RemoteAddr)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Passkey berhasil dihapus"})
}
