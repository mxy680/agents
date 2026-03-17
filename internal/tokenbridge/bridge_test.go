package tokenbridge

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/emdash-projects/agents/internal/portal/crypto"
	"github.com/emdash-projects/agents/internal/portal/database"
)

func TestExportGoogle(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	accessToken, _ := crypto.Encrypt(key, "google-access-tok")
	refreshToken, _ := crypto.Encrypt(key, "google-refresh-tok")

	b := &Bridge{key: key}
	env := make(map[string]string)

	intg := database.Integration{
		Provider:     "google",
		AccessToken:  pgtype.Text{String: accessToken, Valid: true},
		RefreshToken: pgtype.Text{String: refreshToken, Valid: true},
	}

	err := b.exportGoogle(intg, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env["GOOGLE_ACCESS_TOKEN"] != "google-access-tok" {
		t.Errorf("got %q, want %q", env["GOOGLE_ACCESS_TOKEN"], "google-access-tok")
	}
	if env["GOOGLE_REFRESH_TOKEN"] != "google-refresh-tok" {
		t.Errorf("got %q, want %q", env["GOOGLE_REFRESH_TOKEN"], "google-refresh-tok")
	}
}

func TestExportGoogleAccessTokenOnly(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	accessToken, _ := crypto.Encrypt(key, "access-only")

	b := &Bridge{key: key}
	env := make(map[string]string)

	intg := database.Integration{
		Provider:    "google",
		AccessToken: pgtype.Text{String: accessToken, Valid: true},
		// RefreshToken not set
	}

	if err := b.exportGoogle(intg, env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env["GOOGLE_ACCESS_TOKEN"] != "access-only" {
		t.Errorf("got %q, want %q", env["GOOGLE_ACCESS_TOKEN"], "access-only")
	}
	if _, ok := env["GOOGLE_REFRESH_TOKEN"]; ok {
		t.Error("expected GOOGLE_REFRESH_TOKEN to be absent")
	}
}

func TestExportGitHub(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	accessToken, _ := crypto.Encrypt(key, "gh-access-tok")
	refreshToken, _ := crypto.Encrypt(key, "gh-refresh-tok")

	b := &Bridge{key: key}
	env := make(map[string]string)

	intg := database.Integration{
		Provider:     "github",
		AccessToken:  pgtype.Text{String: accessToken, Valid: true},
		RefreshToken: pgtype.Text{String: refreshToken, Valid: true},
	}

	err := b.exportGitHub(intg, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env["GITHUB_ACCESS_TOKEN"] != "gh-access-tok" {
		t.Errorf("got %q, want %q", env["GITHUB_ACCESS_TOKEN"], "gh-access-tok")
	}
	if env["GITHUB_REFRESH_TOKEN"] != "gh-refresh-tok" {
		t.Errorf("got %q, want %q", env["GITHUB_REFRESH_TOKEN"], "gh-refresh-tok")
	}
}

func TestExportInstagram(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	sessionID, _ := crypto.Encrypt(key, "sess123")
	csrfToken, _ := crypto.Encrypt(key, "csrf456")
	dsUserID, _ := crypto.Encrypt(key, "user789")

	meta := map[string]string{
		"session_id": sessionID,
		"csrf_token": csrfToken,
		"ds_user_id": dsUserID,
	}
	metaJSON, _ := json.Marshal(meta)

	b := &Bridge{key: key}
	env := make(map[string]string)

	intg := database.Integration{
		Provider: "instagram",
		Metadata: metaJSON,
	}

	err := b.exportInstagram(intg, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env["INSTAGRAM_SESSION_ID"] != "sess123" {
		t.Errorf("got %q, want %q", env["INSTAGRAM_SESSION_ID"], "sess123")
	}
	if env["INSTAGRAM_CSRF_TOKEN"] != "csrf456" {
		t.Errorf("got %q, want %q", env["INSTAGRAM_CSRF_TOKEN"], "csrf456")
	}
	if env["INSTAGRAM_DS_USER_ID"] != "user789" {
		t.Errorf("got %q, want %q", env["INSTAGRAM_DS_USER_ID"], "user789")
	}
}

func TestExportInstagramWithOptionalFields(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	sessionID, _ := crypto.Encrypt(key, "sess123")
	csrfToken, _ := crypto.Encrypt(key, "csrf456")
	dsUserID, _ := crypto.Encrypt(key, "user789")
	mid, _ := crypto.Encrypt(key, "mid-val")
	igDid, _ := crypto.Encrypt(key, "did-val")

	meta := map[string]string{
		"session_id": sessionID,
		"csrf_token": csrfToken,
		"ds_user_id": dsUserID,
		"mid":        mid,
		"ig_did":     igDid,
	}
	metaJSON, _ := json.Marshal(meta)

	b := &Bridge{key: key}
	env := make(map[string]string)

	intg := database.Integration{
		Provider: "instagram",
		Metadata: metaJSON,
	}

	if err := b.exportInstagram(intg, env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env["INSTAGRAM_MID"] != "mid-val" {
		t.Errorf("got %q, want %q", env["INSTAGRAM_MID"], "mid-val")
	}
	if env["INSTAGRAM_IG_DID"] != "did-val" {
		t.Errorf("got %q, want %q", env["INSTAGRAM_IG_DID"], "did-val")
	}
}

func TestExportInstagramEmpty(t *testing.T) {
	b := &Bridge{key: make([]byte, 32)}
	env := make(map[string]string)

	intg := database.Integration{Provider: "instagram", Metadata: nil}
	if err := b.exportInstagram(intg, env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected empty env, got %v", env)
	}
}
