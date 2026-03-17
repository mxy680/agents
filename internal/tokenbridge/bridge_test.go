package tokenbridge

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// testKey is a deterministic 32-byte key for tests (64 hex chars).
const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// encrypt mirrors the portal's Node.js encrypt function for test fixtures.
func encrypt(plaintext string, hexKey string) string {
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		fmt.Sscanf(hexKey[i*2:i*2+2], "%02x", &key[i])
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple", "hello-world"},
		{"empty", ""},
		{"token-like", "ya29.a0AfH6SMBxG..."},
		{"unicode", "testing-123-\u00e9\u00e8\u00ea"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := encrypt(tt.plaintext, testKey)
			got, err := Decrypt(encoded, testKey)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}
			if got != tt.plaintext {
				t.Errorf("Decrypt() = %q, want %q", got, tt.plaintext)
			}
		})
	}
}

func TestDecryptErrors(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		key     string
	}{
		{"bad base64", "not-base64!!!", testKey},
		{"too short", base64.StdEncoding.EncodeToString([]byte("short")), testKey},
		{"bad key hex", encrypt("test", testKey), "not-hex"},
		{"wrong key length", encrypt("test", testKey), "0123456789abcdef"},
		{"wrong key", encrypt("test", testKey), "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.encoded, tt.key)
			if err == nil {
				t.Fatal("Decrypt() expected error, got nil")
			}
		})
	}
}

func TestProcessIntegrationGoogle(t *testing.T) {
	integ := &Integration{
		Provider:     "google",
		AccessToken:  sql.NullString{String: encrypt("goog-access", testKey), Valid: true},
		RefreshToken: sql.NullString{String: encrypt("goog-refresh", testKey), Valid: true},
	}
	env := make(map[string]string)
	if err := processIntegration(integ, testKey, env); err != nil {
		t.Fatalf("processIntegration(google) error = %v", err)
	}
	if env["GOOGLE_ACCESS_TOKEN"] != "goog-access" {
		t.Errorf("GOOGLE_ACCESS_TOKEN = %q, want %q", env["GOOGLE_ACCESS_TOKEN"], "goog-access")
	}
	if env["GOOGLE_REFRESH_TOKEN"] != "goog-refresh" {
		t.Errorf("GOOGLE_REFRESH_TOKEN = %q, want %q", env["GOOGLE_REFRESH_TOKEN"], "goog-refresh")
	}
}

func TestProcessIntegrationGitHub(t *testing.T) {
	integ := &Integration{
		Provider:     "github",
		AccessToken:  sql.NullString{String: encrypt("gh-token", testKey), Valid: true},
		RefreshToken: sql.NullString{Valid: false},
	}
	env := make(map[string]string)
	if err := processIntegration(integ, testKey, env); err != nil {
		t.Fatalf("processIntegration(github) error = %v", err)
	}
	if env["GITHUB_ACCESS_TOKEN"] != "gh-token" {
		t.Errorf("GITHUB_ACCESS_TOKEN = %q, want %q", env["GITHUB_ACCESS_TOKEN"], "gh-token")
	}
	if _, ok := env["GITHUB_REFRESH_TOKEN"]; ok {
		t.Error("GITHUB_REFRESH_TOKEN should not be set")
	}
}

func TestProcessIntegrationInstagram(t *testing.T) {
	metadata := map[string]string{
		"session_id": encrypt("ig-sess", testKey),
		"csrf_token": encrypt("ig-csrf", testKey),
		"ds_user_id": encrypt("ig-user", testKey),
	}
	raw, _ := json.Marshal(metadata)
	integ := &Integration{
		Provider: "instagram",
		Metadata: raw,
	}
	env := make(map[string]string)
	if err := processIntegration(integ, testKey, env); err != nil {
		t.Fatalf("processIntegration(instagram) error = %v", err)
	}
	if env["INSTAGRAM_SESSION_ID"] != "ig-sess" {
		t.Errorf("INSTAGRAM_SESSION_ID = %q, want %q", env["INSTAGRAM_SESSION_ID"], "ig-sess")
	}
}

func TestProcessIntegrationUnknownProvider(t *testing.T) {
	integ := &Integration{Provider: "slack"}
	env := make(map[string]string)
	if err := processIntegration(integ, testKey, env); err != nil {
		t.Fatalf("processIntegration(unknown) error = %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected empty env for unknown provider, got %v", env)
	}
}

func TestDecryptOAuth(t *testing.T) {
	integ := &Integration{
		AccessToken:  sql.NullString{String: encrypt("my-access-token", testKey), Valid: true},
		RefreshToken: sql.NullString{String: encrypt("my-refresh-token", testKey), Valid: true},
	}

	env := make(map[string]string)
	err := decryptOAuth(integ, testKey, "GOOGLE", env)
	if err != nil {
		t.Fatalf("decryptOAuth() error = %v", err)
	}

	if env["GOOGLE_ACCESS_TOKEN"] != "my-access-token" {
		t.Errorf("GOOGLE_ACCESS_TOKEN = %q, want %q", env["GOOGLE_ACCESS_TOKEN"], "my-access-token")
	}
	if env["GOOGLE_REFRESH_TOKEN"] != "my-refresh-token" {
		t.Errorf("GOOGLE_REFRESH_TOKEN = %q, want %q", env["GOOGLE_REFRESH_TOKEN"], "my-refresh-token")
	}
}

func TestDecryptOAuthNulls(t *testing.T) {
	integ := &Integration{
		AccessToken:  sql.NullString{Valid: false},
		RefreshToken: sql.NullString{Valid: false},
	}

	env := make(map[string]string)
	err := decryptOAuth(integ, testKey, "GITHUB", env)
	if err != nil {
		t.Fatalf("decryptOAuth() error = %v", err)
	}

	if len(env) != 0 {
		t.Errorf("expected empty env map, got %v", env)
	}
}

func TestDecryptInstagram(t *testing.T) {
	metadata := map[string]string{
		"session_id": encrypt("sess-123", testKey),
		"csrf_token": encrypt("csrf-456", testKey),
		"ds_user_id": encrypt("user-789", testKey),
		"mid":        encrypt("mid-abc", testKey),
	}
	raw, _ := json.Marshal(metadata)

	env := make(map[string]string)
	err := decryptInstagram(raw, testKey, env)
	if err != nil {
		t.Fatalf("decryptInstagram() error = %v", err)
	}

	expected := map[string]string{
		"INSTAGRAM_SESSION_ID": "sess-123",
		"INSTAGRAM_CSRF_TOKEN": "csrf-456",
		"INSTAGRAM_DS_USER_ID": "user-789",
		"INSTAGRAM_MID":        "mid-abc",
	}

	for k, want := range expected {
		if env[k] != want {
			t.Errorf("%s = %q, want %q", k, env[k], want)
		}
	}
}

func TestDecryptInstagramEmpty(t *testing.T) {
	env := make(map[string]string)

	// Empty JSON
	err := decryptInstagram(json.RawMessage("{}"), testKey, env)
	if err != nil {
		t.Fatalf("decryptInstagram({}) error = %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected empty env map, got %v", env)
	}

	// Nil
	err = decryptInstagram(nil, testKey, env)
	if err != nil {
		t.Fatalf("decryptInstagram(nil) error = %v", err)
	}
}

func TestDecryptOAuthBadAccessToken(t *testing.T) {
	integ := &Integration{
		AccessToken:  sql.NullString{String: "not-valid-base64!!!", Valid: true},
		RefreshToken: sql.NullString{Valid: false},
	}
	env := make(map[string]string)
	err := decryptOAuth(integ, testKey, "GOOGLE", env)
	if err == nil {
		t.Fatal("expected error for bad access token")
	}
}

func TestDecryptOAuthBadRefreshToken(t *testing.T) {
	integ := &Integration{
		AccessToken:  sql.NullString{Valid: false},
		RefreshToken: sql.NullString{String: "not-valid-base64!!!", Valid: true},
	}
	env := make(map[string]string)
	err := decryptOAuth(integ, testKey, "GOOGLE", env)
	if err == nil {
		t.Fatal("expected error for bad refresh token")
	}
}

func TestDecryptInstagramBadJSON(t *testing.T) {
	env := make(map[string]string)
	err := decryptInstagram(json.RawMessage("{bad-json"), testKey, env)
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestDecryptInstagramBadEncryptedValue(t *testing.T) {
	metadata := map[string]string{
		"session_id": "not-valid-encrypted-data",
	}
	raw, _ := json.Marshal(metadata)
	env := make(map[string]string)
	err := decryptInstagram(raw, testKey, env)
	if err == nil {
		t.Fatal("expected error for bad encrypted value")
	}
}

func TestProcessIntegrationDecryptError(t *testing.T) {
	integ := &Integration{
		Provider:    "google",
		AccessToken: sql.NullString{String: "bad-data", Valid: true},
	}
	env := make(map[string]string)
	err := processIntegration(integ, testKey, env)
	if err == nil {
		t.Fatal("expected error for bad encrypted data")
	}
}

func TestProcessIntegrationGitHubDecryptError(t *testing.T) {
	integ := &Integration{
		Provider:    "github",
		AccessToken: sql.NullString{String: "bad-data", Valid: true},
	}
	env := make(map[string]string)
	err := processIntegration(integ, testKey, env)
	if err == nil {
		t.Fatal("expected error for bad encrypted data")
	}
}

func TestProcessIntegrationInstagramDecryptError(t *testing.T) {
	integ := &Integration{
		Provider: "instagram",
		Metadata: json.RawMessage(`{"session_id":"bad-data"}`),
	}
	env := make(map[string]string)
	err := processIntegration(integ, testKey, env)
	if err == nil {
		t.Fatal("expected error for bad encrypted data")
	}
}

func TestExportEnvForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	igMeta := map[string]string{
		"session_id": encrypt("ig-sess", testKey),
		"csrf_token": encrypt("ig-csrf", testKey),
		"ds_user_id": encrypt("ig-user", testKey),
	}
	igMetaJSON, _ := json.Marshal(igMeta)

	rows := sqlmock.NewRows([]string{"provider", "access_token", "refresh_token", "metadata"}).
		AddRow("google", encrypt("goog-at", testKey), encrypt("goog-rt", testKey), json.RawMessage("{}")).
		AddRow("github", encrypt("gh-at", testKey), nil, json.RawMessage("{}")).
		AddRow("instagram", nil, nil, igMetaJSON)

	mock.ExpectQuery("SELECT provider, access_token, refresh_token, metadata").
		WithArgs("user-123").
		WillReturnRows(rows)

	env, err := ExportEnvForUser(context.Background(), db, "user-123", testKey)
	if err != nil {
		t.Fatalf("ExportEnvForUser() error = %v", err)
	}

	expected := map[string]string{
		"GOOGLE_ACCESS_TOKEN":  "goog-at",
		"GOOGLE_REFRESH_TOKEN": "goog-rt",
		"GITHUB_ACCESS_TOKEN":  "gh-at",
		"INSTAGRAM_SESSION_ID": "ig-sess",
		"INSTAGRAM_CSRF_TOKEN": "ig-csrf",
		"INSTAGRAM_DS_USER_ID": "ig-user",
	}

	for k, want := range expected {
		if env[k] != want {
			t.Errorf("%s = %q, want %q", k, env[k], want)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestExportEnvForUserQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("connection refused"))

	_, err = ExportEnvForUser(context.Background(), db, "user-123", testKey)
	if err == nil {
		t.Fatal("expected error for query failure")
	}
}

func TestExportEnvForUserEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"provider", "access_token", "refresh_token", "metadata"})
	mock.ExpectQuery("SELECT").WithArgs("user-456").WillReturnRows(rows)

	env, err := ExportEnvForUser(context.Background(), db, "user-456", testKey)
	if err != nil {
		t.Fatalf("ExportEnvForUser() error = %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected empty env, got %v", env)
	}
}
