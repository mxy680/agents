package tokenbridge

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// testKey is a deterministic 32-byte key for tests (64 hex chars).
const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// encryptBytes mirrors the portal's encryptToBytes for test fixtures.
func encryptBytes(plaintext string, hexKey string) []byte {
	key, _ := hex.DecodeString(hexKey)
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

	return gcm.Seal(nonce, nonce, []byte(plaintext), nil)
}

// encryptCredentialsForTest encrypts a credential map to bytea.
func encryptCredentialsForTest(creds map[string]string) []byte {
	raw, _ := json.Marshal(creds)
	return encryptBytes(string(raw), testKey)
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
			encrypted := encryptBytes(tt.plaintext, testKey)
			got, err := Decrypt(encrypted, testKey)
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
		name string
		data []byte
		key  string
	}{
		{"too short", []byte("short"), testKey},
		{"bad key hex", encryptBytes("test", testKey), "not-hex"},
		{"wrong key length", encryptBytes("test", testKey), "0123456789abcdef"},
		{"wrong key", encryptBytes("test", testKey), "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt(tt.data, tt.key)
			if err == nil {
				t.Fatal("Decrypt() expected error, got nil")
			}
		})
	}
}

func TestDecryptCredentials(t *testing.T) {
	creds := map[string]string{
		"access_token":  "my-token",
		"refresh_token": "my-refresh",
	}
	encrypted := encryptCredentialsForTest(creds)

	got, err := DecryptCredentials(encrypted, testKey)
	if err != nil {
		t.Fatalf("DecryptCredentials() error = %v", err)
	}
	if got["access_token"] != "my-token" {
		t.Errorf("access_token = %q, want %q", got["access_token"], "my-token")
	}
	if got["refresh_token"] != "my-refresh" {
		t.Errorf("refresh_token = %q, want %q", got["refresh_token"], "my-refresh")
	}
}

func TestDecryptCredentialsBadJSON(t *testing.T) {
	encrypted := encryptBytes("not-json", testKey)
	_, err := DecryptCredentials(encrypted, testKey)
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestProcessIntegrationGoogle(t *testing.T) {
	creds := encryptCredentialsForTest(map[string]string{
		"access_token":  "goog-at",
		"refresh_token": "goog-rt",
	})
	ui := &UserIntegration{Provider: "google", Credentials: creds}
	env := make(map[string]string)
	if err := processIntegration(ui, testKey, env); err != nil {
		t.Fatalf("error = %v", err)
	}
	if env["GOOGLE_ACCESS_TOKEN"] != "goog-at" {
		t.Errorf("GOOGLE_ACCESS_TOKEN = %q", env["GOOGLE_ACCESS_TOKEN"])
	}
	if env["GOOGLE_REFRESH_TOKEN"] != "goog-rt" {
		t.Errorf("GOOGLE_REFRESH_TOKEN = %q", env["GOOGLE_REFRESH_TOKEN"])
	}
}

func TestProcessIntegrationGitHub(t *testing.T) {
	creds := encryptCredentialsForTest(map[string]string{
		"access_token": "gh-at",
	})
	ui := &UserIntegration{Provider: "github", Credentials: creds}
	env := make(map[string]string)
	if err := processIntegration(ui, testKey, env); err != nil {
		t.Fatalf("error = %v", err)
	}
	if env["GITHUB_ACCESS_TOKEN"] != "gh-at" {
		t.Errorf("GITHUB_ACCESS_TOKEN = %q", env["GITHUB_ACCESS_TOKEN"])
	}
	if _, ok := env["GITHUB_REFRESH_TOKEN"]; ok {
		t.Error("GITHUB_REFRESH_TOKEN should not be set")
	}
}

func TestProcessIntegrationInstagram(t *testing.T) {
	creds := encryptCredentialsForTest(map[string]string{
		"session_id": "ig-sess",
		"csrf_token": "ig-csrf",
		"ds_user_id": "ig-user",
		"mid":        "ig-mid",
	})
	ui := &UserIntegration{Provider: "instagram", Credentials: creds}
	env := make(map[string]string)
	if err := processIntegration(ui, testKey, env); err != nil {
		t.Fatalf("error = %v", err)
	}
	expected := map[string]string{
		"INSTAGRAM_SESSION_ID": "ig-sess",
		"INSTAGRAM_CSRF_TOKEN": "ig-csrf",
		"INSTAGRAM_DS_USER_ID": "ig-user",
		"INSTAGRAM_MID":        "ig-mid",
	}
	for k, want := range expected {
		if env[k] != want {
			t.Errorf("%s = %q, want %q", k, env[k], want)
		}
	}
}

func TestProcessIntegrationLinkedIn(t *testing.T) {
	creds := encryptCredentialsForTest(map[string]string{
		"li_at":      "li-at-token",
		"jsessionid": "ajax:1234567890",
	})
	ui := &UserIntegration{Provider: "linkedin", Credentials: creds}
	env := make(map[string]string)
	if err := processIntegration(ui, testKey, env); err != nil {
		t.Fatalf("error = %v", err)
	}
	expected := map[string]string{
		"LINKEDIN_LI_AT":     "li-at-token",
		"LINKEDIN_JSESSIONID": "ajax:1234567890",
	}
	for k, want := range expected {
		if env[k] != want {
			t.Errorf("%s = %q, want %q", k, env[k], want)
		}
	}
}

func TestProcessIntegrationUnknown(t *testing.T) {
	creds := encryptCredentialsForTest(map[string]string{"key": "val"})
	ui := &UserIntegration{Provider: "slack", Credentials: creds}
	env := make(map[string]string)
	if err := processIntegration(ui, testKey, env); err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected empty env, got %v", env)
	}
}

func TestProcessIntegrationDecryptError(t *testing.T) {
	ui := &UserIntegration{Provider: "google", Credentials: []byte("bad")}
	env := make(map[string]string)
	err := processIntegration(ui, testKey, env)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExportEnvForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	googleCreds := encryptCredentialsForTest(map[string]string{
		"access_token": "goog-at", "refresh_token": "goog-rt",
	})
	ghCreds := encryptCredentialsForTest(map[string]string{
		"access_token": "gh-at",
	})
	igCreds := encryptCredentialsForTest(map[string]string{
		"session_id": "ig-sess", "csrf_token": "ig-csrf", "ds_user_id": "ig-user",
	})

	rows := sqlmock.NewRows([]string{"provider", "credentials"}).
		AddRow("google", googleCreds).
		AddRow("github", ghCreds).
		AddRow("instagram", igCreds)

	mock.ExpectQuery("SELECT").WithArgs("user-123").WillReturnRows(rows)

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
		t.Fatal("expected error")
	}
}

func TestExportEnvForUserEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"provider", "credentials"})
	mock.ExpectQuery("SELECT").WithArgs("user-456").WillReturnRows(rows)

	env, err := ExportEnvForUser(context.Background(), db, "user-456", testKey)
	if err != nil {
		t.Fatalf("ExportEnvForUser() error = %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected empty env, got %v", env)
	}
}
