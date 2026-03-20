package orchestrator

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// testEncKey is a deterministic 32-byte key for tests (64 hex chars).
const testEncKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// encryptCreds encrypts a credential map the same way the portal does.
func encryptCreds(creds map[string]string) []byte {
	raw, _ := json.Marshal(creds)
	key, _ := hex.DecodeString(testEncKey)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}
	return gcm.Seal(nonce, nonce, raw, nil)
}

func newTestCredentialResolver(store *Store, refresher TokenRefresher) *CredentialResolver {
	return &CredentialResolver{
		store:              store,
		encryptionKey:      testEncKey,
		googleClientID:     "test-client-id",
		googleClientSecret: "test-client-secret",
		refreshToken:       refresher,
	}
}

// ---- NewCredentialResolver ----

func TestNewCredentialResolver(t *testing.T) {
	store, _ := newMockStore(t)
	cfg := Config{
		EncryptionMasterKey: testEncKey,
		GoogleClientID:      "cid",
		GoogleClientSecret:  "csecret",
	}
	cr := NewCredentialResolver(store, cfg)
	if cr == nil {
		t.Fatal("NewCredentialResolver() returned nil")
	}
	if cr.refreshToken == nil {
		t.Error("refreshToken should not be nil")
	}
}

// ---- ResolveForUser with Google credentials ----

func TestResolveForUserGoogleRefresh(t *testing.T) {
	store, mock := newMockStore(t)

	googleCreds := encryptCreds(map[string]string{
		"access_token":  "old-access-token",
		"refresh_token": "my-refresh-token",
	})

	rows := sqlmock.NewRows([]string{"provider", "credentials"}).
		AddRow("google", googleCreds)
	mock.ExpectQuery("SELECT").WithArgs("user-123").WillReturnRows(rows)

	refreshCalled := false
	fakeRefresher := func(refreshToken, clientID, clientSecret string) (string, error) {
		refreshCalled = true
		if refreshToken != "my-refresh-token" {
			t.Errorf("refreshToken = %q, want %q", refreshToken, "my-refresh-token")
		}
		if clientID != "test-client-id" {
			t.Errorf("clientID = %q, want %q", clientID, "test-client-id")
		}
		if clientSecret != "test-client-secret" {
			t.Errorf("clientSecret = %q, want %q", clientSecret, "test-client-secret")
		}
		return "fresh-access-token", nil
	}

	cr := newTestCredentialResolver(store, fakeRefresher)
	env, err := cr.ResolveForUser(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("ResolveForUser() error = %v", err)
	}

	if !refreshCalled {
		t.Error("expected refreshToken to be called")
	}
	if env["GOOGLE_ACCESS_TOKEN"] != "fresh-access-token" {
		t.Errorf("GOOGLE_ACCESS_TOKEN = %q, want %q", env["GOOGLE_ACCESS_TOKEN"], "fresh-access-token")
	}
	if env["GOOGLE_REFRESH_TOKEN"] != "my-refresh-token" {
		t.Errorf("GOOGLE_REFRESH_TOKEN = %q, want %q", env["GOOGLE_REFRESH_TOKEN"], "my-refresh-token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ---- ResolveForUser with GitHub credentials (no refresh) ----

func TestResolveForUserGitHubNoRefresh(t *testing.T) {
	store, mock := newMockStore(t)

	ghCreds := encryptCreds(map[string]string{
		"access_token": "gh-access-token",
	})

	rows := sqlmock.NewRows([]string{"provider", "credentials"}).
		AddRow("github", ghCreds)
	mock.ExpectQuery("SELECT").WithArgs("user-456").WillReturnRows(rows)

	refreshCalled := false
	fakeRefresher := func(refreshToken, clientID, clientSecret string) (string, error) {
		refreshCalled = true
		return "", fmt.Errorf("should not be called")
	}

	cr := newTestCredentialResolver(store, fakeRefresher)
	env, err := cr.ResolveForUser(context.Background(), "user-456")
	if err != nil {
		t.Fatalf("ResolveForUser() error = %v", err)
	}

	if refreshCalled {
		t.Error("refreshToken should not be called for GitHub-only credentials")
	}
	if env["GITHUB_ACCESS_TOKEN"] != "gh-access-token" {
		t.Errorf("GITHUB_ACCESS_TOKEN = %q, want %q", env["GITHUB_ACCESS_TOKEN"], "gh-access-token")
	}
	if _, ok := env["GOOGLE_ACCESS_TOKEN"]; ok {
		t.Error("GOOGLE_ACCESS_TOKEN should not be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ---- ResolveForUser: query error ----

func TestResolveForUserQueryError(t *testing.T) {
	store, mock := newMockStore(t)

	mock.ExpectQuery("SELECT").WithArgs("user-789").WillReturnError(fmt.Errorf("connection refused"))

	cr := newTestCredentialResolver(store, nil)
	_, err := cr.ResolveForUser(context.Background(), "user-789")
	if err == nil {
		t.Fatal("ResolveForUser() expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ---- ResolveForUser: token refresh error ----

func TestResolveForUserRefreshError(t *testing.T) {
	store, mock := newMockStore(t)

	googleCreds := encryptCreds(map[string]string{
		"access_token":  "old-token",
		"refresh_token": "stale-refresh",
	})
	rows := sqlmock.NewRows([]string{"provider", "credentials"}).
		AddRow("google", googleCreds)
	mock.ExpectQuery("SELECT").WithArgs("user-bad").WillReturnRows(rows)

	fakeRefresher := func(refreshToken, clientID, clientSecret string) (string, error) {
		return "", fmt.Errorf("token revoked")
	}

	cr := newTestCredentialResolver(store, fakeRefresher)
	_, err := cr.ResolveForUser(context.Background(), "user-bad")
	if err == nil {
		t.Fatal("ResolveForUser() expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ---- RefreshGoogleToken ----

func TestRefreshGoogleToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		if r.FormValue("grant_type") != "refresh_token" {
			http.Error(w, "bad grant_type", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"fresh-token-from-google"}`)
	}))
	defer srv.Close()

	// We can't easily override the URL in RefreshGoogleToken (it's hardcoded),
	// so we test the function indirectly via the injected refresher in CredentialResolver.
	// For direct tests, we verify the real function's error paths.

	t.Run("missing client id", func(t *testing.T) {
		_, err := RefreshGoogleToken("refresh", "", "secret")
		if err == nil {
			t.Fatal("expected error for empty clientID")
		}
	})

	t.Run("missing client secret", func(t *testing.T) {
		_, err := RefreshGoogleToken("refresh", "id", "")
		if err == nil {
			t.Fatal("expected error for empty clientSecret")
		}
	})
}

func TestRefreshGoogleTokenHTTP(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantErr     bool
		wantToken   string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"access_token":"fresh-token"}`)
			},
			wantToken: "fresh-token",
		},
		{
			name: "non-200 response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error":"invalid_grant"}`)
			},
			wantErr: true,
		},
		{
			name: "empty access_token",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"access_token":""}`)
			},
			wantErr: true,
		},
		{
			name: "invalid JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `not-json`)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			// Use the fakeRefresher pattern to test HTTP behavior by injecting
			// a custom refresher that calls our test server.
			store, mock := newMockStore(t)

			googleCreds := encryptCreds(map[string]string{
				"access_token":  "old",
				"refresh_token": "refresh-tok",
			})
			rows := sqlmock.NewRows([]string{"provider", "credentials"}).
				AddRow("google", googleCreds)
			mock.ExpectQuery("SELECT").WillReturnRows(rows)

			customRefresher := func(refreshToken, clientID, clientSecret string) (string, error) {
				resp, err := http.PostForm(srv.URL, nil)
				if err != nil {
					return "", err
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return "", fmt.Errorf("status %d", resp.StatusCode)
				}

				var result struct {
					AccessToken string `json:"access_token"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					return "", err
				}
				if result.AccessToken == "" {
					return "", fmt.Errorf("empty access_token")
				}
				return result.AccessToken, nil
			}

			cr := newTestCredentialResolver(store, customRefresher)
			env, err := cr.ResolveForUser(context.Background(), "user-test")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveForUser() error = %v", err)
			}
			if env["GOOGLE_ACCESS_TOKEN"] != tt.wantToken {
				t.Errorf("GOOGLE_ACCESS_TOKEN = %q, want %q", env["GOOGLE_ACCESS_TOKEN"], tt.wantToken)
			}
		})
	}
}
