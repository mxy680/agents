package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/emdash-projects/agents/internal/tokenbridge"
)

// TokenRefresher is a function that refreshes a Google access token.
type TokenRefresher func(refreshToken, clientID, clientSecret string) (string, error)

// CredentialResolver resolves credentials for a user's integrations.
type CredentialResolver struct {
	store              *Store
	encryptionKey      string
	googleClientID     string
	googleClientSecret string
	refreshToken       TokenRefresher
}

// NewCredentialResolver creates a new CredentialResolver.
func NewCredentialResolver(store *Store, cfg Config) *CredentialResolver {
	return &CredentialResolver{
		store:              store,
		encryptionKey:      cfg.EncryptionMasterKey,
		googleClientID:     cfg.GoogleClientID,
		googleClientSecret: cfg.GoogleClientSecret,
		refreshToken:       RefreshGoogleToken,
	}
}

// ResolveForUser returns all credential env vars for a user, with fresh Google tokens.
func (cr *CredentialResolver) ResolveForUser(ctx context.Context, userID string) (map[string]string, error) {
	env, err := tokenbridge.ExportEnvForUser(ctx, cr.store.DB(), userID, cr.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("export env: %w", err)
	}

	// For Google: refresh the access token
	if refreshToken, ok := env["GOOGLE_REFRESH_TOKEN"]; ok && refreshToken != "" {
		freshToken, err := cr.refreshToken(refreshToken, cr.googleClientID, cr.googleClientSecret)
		if err != nil {
			return nil, fmt.Errorf("refresh google token: %w", err)
		}
		env["GOOGLE_ACCESS_TOKEN"] = freshToken
	}

	return env, nil
}

// RefreshGoogleToken exchanges a refresh token for a fresh access token.
func RefreshGoogleToken(refreshToken, clientID, clientSecret string) (string, error) {
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	})
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, body)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
	}

	return result.AccessToken, nil
}
