package tokenbridge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Integration represents a row from the integrations table.
type Integration struct {
	Provider     string
	AccessToken  sql.NullString
	RefreshToken sql.NullString
	Metadata     json.RawMessage
}

// DB abstracts the database operations needed by the bridge.
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// ExportEnvForUser reads all active integrations for a user and returns
// a map of environment variable names to decrypted values.
func ExportEnvForUser(ctx context.Context, db DB, userID string, hexKey string) (map[string]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT provider, access_token, refresh_token, metadata
		 FROM integrations
		 WHERE user_id = $1 AND status = 'active'`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query integrations: %w", err)
	}
	defer rows.Close()

	env := make(map[string]string)
	for rows.Next() {
		var integ Integration
		if err := rows.Scan(&integ.Provider, &integ.AccessToken, &integ.RefreshToken, &integ.Metadata); err != nil {
			return nil, fmt.Errorf("scan integration: %w", err)
		}

		if err := processIntegration(&integ, hexKey, env); err != nil {
			return nil, err
		}
	}

	return env, rows.Err()
}

// processIntegration decrypts a single integration's tokens into env vars.
func processIntegration(integ *Integration, hexKey string, env map[string]string) error {
	switch integ.Provider {
	case "google":
		if err := decryptOAuth(integ, hexKey, "GOOGLE", env); err != nil {
			return fmt.Errorf("decrypt google: %w", err)
		}
	case "github":
		if err := decryptOAuth(integ, hexKey, "GITHUB", env); err != nil {
			return fmt.Errorf("decrypt github: %w", err)
		}
	case "instagram":
		if err := decryptInstagram(integ.Metadata, hexKey, env); err != nil {
			return fmt.Errorf("decrypt instagram: %w", err)
		}
	}
	return nil
}

func decryptOAuth(integ *Integration, hexKey string, prefix string, env map[string]string) error {
	if integ.AccessToken.Valid {
		val, err := Decrypt(integ.AccessToken.String, hexKey)
		if err != nil {
			return fmt.Errorf("access_token: %w", err)
		}
		env[prefix+"_ACCESS_TOKEN"] = val
	}
	if integ.RefreshToken.Valid {
		val, err := Decrypt(integ.RefreshToken.String, hexKey)
		if err != nil {
			return fmt.Errorf("refresh_token: %w", err)
		}
		env[prefix+"_REFRESH_TOKEN"] = val
	}
	return nil
}

var instagramKeyMap = map[string]string{
	"session_id": "INSTAGRAM_SESSION_ID",
	"csrf_token": "INSTAGRAM_CSRF_TOKEN",
	"ds_user_id": "INSTAGRAM_DS_USER_ID",
	"mid":        "INSTAGRAM_MID",
	"ig_did":     "INSTAGRAM_IG_DID",
}

func decryptInstagram(raw json.RawMessage, hexKey string, env map[string]string) error {
	if len(raw) == 0 || string(raw) == "{}" {
		return nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return fmt.Errorf("unmarshal metadata: %w", err)
	}

	for jsonKey, envVar := range instagramKeyMap {
		encrypted, ok := metadata[jsonKey]
		if !ok || encrypted == "" {
			continue
		}
		val, err := Decrypt(encrypted, hexKey)
		if err != nil {
			return fmt.Errorf("%s: %w", jsonKey, err)
		}
		env[envVar] = val
	}

	return nil
}
