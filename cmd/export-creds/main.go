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
		// construct from known parts
		dbURL = fmt.Sprintf("postgres://postgres:%s@db.juetvofnwfjylyfgqbvt.supabase.co:5432/postgres?sslmode=require",
			os.Getenv("SUPABASE_DB_PASSWORD"))
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
		SELECT ui.credentials
		FROM user_integrations ui
		JOIN integrations i ON i.id = ui.integration_id
		WHERE ui.user_id = $1 AND i.name = $2 AND ui.status = 'connected'`
	args := []any{*userID, *provider}

	if *account != "" {
		query += " AND ui.account_label = $3"
		args = append(args, *account)
	}
	query += " LIMIT 1"

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

	// Env var mappings per provider
	mappings := map[string]map[string]string{
		"google": {
			"access_token":  "GOOGLE_ACCESS_TOKEN",
			"refresh_token": "GOOGLE_REFRESH_TOKEN",
		},
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
		// unknown provider: just dump raw JSON
		_ = json.NewEncoder(os.Stdout).Encode(creds)
		return
	}

	for credKey, envVar := range m {
		if val, ok := creds[credKey]; ok && val != "" {
			fmt.Printf("export %s=%q\n", envVar, val)
		}
	}
}
