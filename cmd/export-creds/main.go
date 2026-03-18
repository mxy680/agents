// export-creds prints shell export statements for a user's integration credentials.
// Usage: export-creds --user-id UUID --provider google [--account personal]
// Example: eval $(export-creds --user-id abc --provider google --account personal)
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/emdash-projects/agents/internal/tokenbridge"
	_ "github.com/lib/pq"
)

func main() {
	userID := flag.String("user-id", "", "User UUID")
	provider := flag.String("provider", "", "Provider name (google, github, instagram)")
	account := flag.String("account", "", "Account label (empty = first found)")
	flag.Parse()

	if *userID == "" || *provider == "" {
		fmt.Fprintln(os.Stderr, "Usage: export-creds --user-id UUID --provider PROVIDER [--account LABEL]")
		os.Exit(1)
	}

	dbURL := os.Getenv("SUPABASE_DB_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "SUPABASE_DB_URL not set")
		os.Exit(1)
	}

	encKey := os.Getenv("ENCRYPTION_MASTER_KEY")
	if encKey == "" {
		fmt.Fprintln(os.Stderr, "ENCRYPTION_MASTER_KEY not set")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db open:", err)
		os.Exit(1)
	}
	defer db.Close()

	query := `
		SELECT credentials
		FROM user_integrations
		WHERE user_id = $1 AND provider = $2 AND status = 'active'`
	args := []any{*userID, *provider}

	if *account != "" {
		query += " AND label = $3"
		args = append(args, *account)
	}
	query += " ORDER BY updated_at DESC LIMIT 1"

	var encrypted []byte
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&encrypted); err != nil {
		fmt.Fprintln(os.Stderr, "query:", err)
		os.Exit(1)
	}

	creds, err := tokenbridge.DecryptCredentials(encrypted, encKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "decrypt:", err)
		os.Exit(1)
	}

	// For Google: always mint a fresh access token using the refresh token.
	// The stored access_token is ignored — it expires in 1 hour and is never updated.
	if *provider == "google" {
		refreshToken := creds["refresh_token"]
		if refreshToken == "" {
			fmt.Fprintln(os.Stderr, "no refresh_token found for google")
			os.Exit(1)
		}

		freshToken, err := refreshGoogleToken(refreshToken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "google token refresh:", err)
			os.Exit(1)
		}

		fmt.Printf("export GOOGLE_ACCESS_TOKEN=%q\n", freshToken)
		fmt.Printf("export GOOGLE_REFRESH_TOKEN=%q\n", refreshToken)
		return
	}

	// Env var mappings for non-Google providers
	mappings := map[string]map[string]string{
		"github": {
			"access_token":  "GITHUB_ACCESS_TOKEN",
			"refresh_token": "GITHUB_REFRESH_TOKEN",
		},
		"instagram": {
			"session_id": "INSTAGRAM_SESSION_ID",
			"csrf_token": "INSTAGRAM_CSRF_TOKEN",
			"ds_user_id": "INSTAGRAM_DS_USER_ID",
			"mid":        "INSTAGRAM_MID",
			"ig_did":     "INSTAGRAM_IG_DID",
		},
	}

	m, ok := mappings[*provider]
	if !ok {
		_ = json.NewEncoder(os.Stdout).Encode(creds)
		return
	}

	for credKey, envVar := range m {
		if val, ok := creds[credKey]; ok && val != "" {
			fmt.Printf("export %s=%q\n", envVar, val)
		}
	}
}

// refreshGoogleToken exchanges a refresh token for a fresh access token.
func refreshGoogleToken(refreshToken string) (string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
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
