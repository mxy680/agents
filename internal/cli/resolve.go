package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"github.com/emdash-projects/agents/internal/orchestrator"
	"github.com/emdash-projects/agents/internal/tokenbridge"
)

// resolveUserCredentials is a PersistentPreRunE hook that resolves credentials
// from the user_integrations table when RESOLVE_USER_ID is set.
func resolveUserCredentials(cmd *cobra.Command, _ []string) error {
	userID := os.Getenv("RESOLVE_USER_ID")
	if userID == "" {
		return nil
	}

	dbURL := os.Getenv("SUPABASE_DB_URL")
	if dbURL == "" {
		return fmt.Errorf("RESOLVE_USER_ID is set but SUPABASE_DB_URL is missing")
	}
	encKey := os.Getenv("ENCRYPTION_MASTER_KEY")
	if encKey == "" {
		return fmt.Errorf("RESOLVE_USER_ID is set but ENCRYPTION_MASTER_KEY is missing")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	env, err := tokenbridge.ExportEnvForUser(cmd.Context(), db, userID, encKey)
	if err != nil {
		return fmt.Errorf("resolve credentials: %w", err)
	}

	// Set env vars, but don't override explicitly set ones
	for k, v := range env {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}

	// Refresh Google token if resolved
	if rt := os.Getenv("GOOGLE_REFRESH_TOKEN"); rt != "" {
		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
		if clientID != "" && clientSecret != "" {
			fresh, err := orchestrator.RefreshGoogleToken(rt, clientID, clientSecret)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: google token refresh failed: %v\n", err)
			} else {
				os.Setenv("GOOGLE_ACCESS_TOKEN", fresh)
			}
		}
	}

	// Refresh Supabase token if resolved
	if rt := os.Getenv("SUPABASE_REFRESH_TOKEN"); rt != "" {
		clientID := os.Getenv("SUPABASE_INTEGRATION_CLIENT_ID")
		clientSecret := os.Getenv("SUPABASE_INTEGRATION_CLIENT_SECRET")
		if clientID != "" && clientSecret != "" {
			fresh, err := refreshSupabaseToken(rt, clientID, clientSecret)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: supabase token refresh failed: %v\n", err)
			} else {
				os.Setenv("SUPABASE_ACCESS_TOKEN", fresh)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "resolved %d credential(s) for user %s\n", len(env), userID)
	return nil
}

// refreshSupabaseToken exchanges a Supabase refresh token for a fresh access token.
func refreshSupabaseToken(refreshToken, clientID, clientSecret string) (string, error) {
	resp, err := http.PostForm("https://api.supabase.com/v1/oauth/token", url.Values{
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
