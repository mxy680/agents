// Package tokenbridge reads encrypted tokens from the database and exports them
// as environment variable maps for the CLI binary.
package tokenbridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/emdash-projects/agents/internal/portal/crypto"
	"github.com/emdash-projects/agents/internal/portal/database"
)

// Bridge reads encrypted tokens from the database and exports them
// as environment variables for the CLI binary.
type Bridge struct {
	queries *database.Queries
	key     []byte
}

// NewBridge creates a new token bridge.
func NewBridge(queries *database.Queries, encryptionKey []byte) *Bridge {
	return &Bridge{queries: queries, key: encryptionKey}
}

// ExportEnvForUser fetches all integrations for a user and returns
// a map of environment variable names to decrypted values.
func (b *Bridge) ExportEnvForUser(ctx context.Context, userID pgtype.UUID) (map[string]string, error) {
	integrations, err := b.queries.GetIntegrationsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch integrations: %w", err)
	}

	env := make(map[string]string)

	for _, intg := range integrations {
		switch intg.Provider {
		case "google":
			if err := b.exportGoogle(intg, env); err != nil {
				return nil, fmt.Errorf("export google: %w", err)
			}
		case "github":
			if err := b.exportGitHub(intg, env); err != nil {
				return nil, fmt.Errorf("export github: %w", err)
			}
		case "instagram":
			if err := b.exportInstagram(intg, env); err != nil {
				return nil, fmt.Errorf("export instagram: %w", err)
			}
		}
	}

	return env, nil
}

func (b *Bridge) exportGoogle(intg database.Integration, env map[string]string) error {
	if intg.AccessToken.Valid {
		tok, err := crypto.Decrypt(b.key, intg.AccessToken.String)
		if err != nil {
			return fmt.Errorf("decrypt access token: %w", err)
		}
		env["GOOGLE_ACCESS_TOKEN"] = tok
	}
	if intg.RefreshToken.Valid {
		tok, err := crypto.Decrypt(b.key, intg.RefreshToken.String)
		if err != nil {
			return fmt.Errorf("decrypt refresh token: %w", err)
		}
		env["GOOGLE_REFRESH_TOKEN"] = tok
	}
	return nil
}

func (b *Bridge) exportGitHub(intg database.Integration, env map[string]string) error {
	if intg.AccessToken.Valid {
		tok, err := crypto.Decrypt(b.key, intg.AccessToken.String)
		if err != nil {
			return fmt.Errorf("decrypt access token: %w", err)
		}
		env["GITHUB_ACCESS_TOKEN"] = tok
	}
	if intg.RefreshToken.Valid {
		tok, err := crypto.Decrypt(b.key, intg.RefreshToken.String)
		if err != nil {
			return fmt.Errorf("decrypt refresh token: %w", err)
		}
		env["GITHUB_REFRESH_TOKEN"] = tok
	}
	return nil
}

func (b *Bridge) exportInstagram(intg database.Integration, env map[string]string) error {
	if len(intg.Metadata) == 0 {
		return nil
	}

	var meta map[string]string
	if err := json.Unmarshal(intg.Metadata, &meta); err != nil {
		return fmt.Errorf("unmarshal metadata: %w", err)
	}

	envMap := map[string]string{
		"session_id": "INSTAGRAM_SESSION_ID",
		"csrf_token": "INSTAGRAM_CSRF_TOKEN",
		"ds_user_id": "INSTAGRAM_DS_USER_ID",
		"mid":        "INSTAGRAM_MID",
		"ig_did":     "INSTAGRAM_IG_DID",
	}

	for metaKey, envKey := range envMap {
		encrypted, ok := meta[metaKey]
		if !ok || encrypted == "" {
			continue
		}
		val, err := crypto.Decrypt(b.key, encrypted)
		if err != nil {
			return fmt.Errorf("decrypt %s: %w", metaKey, err)
		}
		if val != "" {
			env[envKey] = val
		}
	}

	return nil
}
