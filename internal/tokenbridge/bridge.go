package tokenbridge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// UserIntegration represents a row from user_integrations.
type UserIntegration struct {
	Provider    string
	Credentials []byte // encrypted bytea
}

// DB abstracts the database operations needed by the bridge.
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// ExportEnvForUser reads all connected integrations for a user and returns
// a map of environment variable names to decrypted values.
func ExportEnvForUser(ctx context.Context, db DB, userID string, hexKey string) (map[string]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT DISTINCT ON (provider) provider, credentials
		 FROM user_integrations
		 WHERE user_id = $1 AND status = 'active'
		 ORDER BY provider, updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query user_integrations: %w", err)
	}
	defer rows.Close()

	env := make(map[string]string)
	for rows.Next() {
		var ui UserIntegration
		if err := rows.Scan(&ui.Provider, &ui.Credentials); err != nil {
			return nil, fmt.Errorf("scan user_integration: %w", err)
		}

		if err := processIntegration(&ui, hexKey, env); err != nil {
			return nil, err
		}
	}

	return env, rows.Err()
}

// processIntegration decrypts a single integration's credentials into env vars.
func processIntegration(ui *UserIntegration, hexKey string, env map[string]string) error {
	creds, err := DecryptCredentials(ui.Credentials, hexKey)
	if err != nil {
		return fmt.Errorf("decrypt %s credentials: %w", ui.Provider, err)
	}

	switch ui.Provider {
	case "google":
		mapCredentials(creds, env, map[string]string{
			"access_token":  "GOOGLE_ACCESS_TOKEN",
			"refresh_token": "GOOGLE_REFRESH_TOKEN",
		})
	case "github":
		mapCredentials(creds, env, map[string]string{
			"access_token":  "GITHUB_ACCESS_TOKEN",
			"refresh_token": "GITHUB_REFRESH_TOKEN",
		})
	case "instagram":
		mapCredentials(creds, env, map[string]string{
			"session_id": "INSTAGRAM_SESSION_ID",
			"csrf_token": "INSTAGRAM_CSRF_TOKEN",
			"ds_user_id": "INSTAGRAM_DS_USER_ID",
			"mid":        "INSTAGRAM_MID",
			"ig_did":     "INSTAGRAM_IG_DID",
		})
	case "linkedin":
		mapCredentials(creds, env, map[string]string{
			"li_at":      "LINKEDIN_LI_AT",
			"jsessionid": "LINKEDIN_JSESSIONID",
			"bcookie":    "LINKEDIN_BCOOKIE",
			"lidc":       "LINKEDIN_LIDC",
			"li_mc":      "LINKEDIN_LI_MC",
		})
	case "framer":
		mapCredentials(creds, env, map[string]string{
			"api_key":     "FRAMER_API_KEY",
			"project_url": "FRAMER_PROJECT_URL",
		})
	case "supabase":
		mapCredentials(creds, env, map[string]string{
			"access_token":  "SUPABASE_ACCESS_TOKEN",
			"refresh_token": "SUPABASE_REFRESH_TOKEN",
		})
	case "x":
		mapCredentials(creds, env, map[string]string{
			"auth_token": "X_AUTH_TOKEN",
			"csrf_token": "X_CSRF_TOKEN",
		})
	case "bluebubbles":
		mapCredentials(creds, env, map[string]string{
			"url":      "BLUEBUBBLES_URL",
			"password": "BLUEBUBBLES_PASSWORD",
		})
	case "canvas":
		mapCredentials(creds, env, map[string]string{
			"base_url": "CANVAS_BASE_URL",
			"cookies":  "CANVAS_COOKIES",
		})
	case "zillow":
		mapCredentials(creds, env, map[string]string{
			"all_cookies": "ZILLOW_COOKIES",
			"proxy_url":   "ZILLOW_PROXY_URL",
		})
	case "vercel":
		mapCredentials(creds, env, map[string]string{
			"token":   "VERCEL_TOKEN",
			"team_id": "VERCEL_TEAM_ID",
		})
	case "cloudflare":
		mapCredentials(creds, env, map[string]string{
			"token":      "CLOUDFLARE_API_TOKEN",
			"account_id": "CLOUDFLARE_ACCOUNT_ID",
		})
	case "linear":
		mapCredentials(creds, env, map[string]string{
			"token": "LINEAR_API_KEY",
		})
	case "fly":
		mapCredentials(creds, env, map[string]string{
			"token": "FLY_API_TOKEN",
		})
	case "gcp":
		mapCredentials(creds, env, map[string]string{
			"token":      "GCP_ACCESS_TOKEN",
			"project_id": "GCP_PROJECT_ID",
		})
	}
	return nil
}

// DecryptCredentials decrypts a bytea blob into a JSON credential map.
func DecryptCredentials(encrypted []byte, hexKey string) (map[string]string, error) {
	plaintext, err := Decrypt(encrypted, hexKey)
	if err != nil {
		return nil, err
	}

	var creds map[string]string
	if err := json.Unmarshal([]byte(plaintext), &creds); err != nil {
		return nil, fmt.Errorf("unmarshal credentials: %w", err)
	}
	return creds, nil
}

func mapCredentials(creds map[string]string, env map[string]string, mapping map[string]string) {
	for credKey, envVar := range mapping {
		if val, ok := creds[credKey]; ok && val != "" {
			env[envVar] = val
		}
	}
}
