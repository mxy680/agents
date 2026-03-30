package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveCredentials_NoOp(t *testing.T) {
	os.Unsetenv("RESOLVE_CREDENTIALS")

	cmd := &cobra.Command{}
	if err := resolveCredentials(cmd, nil); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestResolveCredentials_MissingDBURL(t *testing.T) {
	t.Setenv("RESOLVE_CREDENTIALS", "1")
	os.Unsetenv("SUPABASE_DB_URL")

	cmd := &cobra.Command{}
	err := resolveCredentials(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing SUPABASE_DB_URL")
	}
	if got := err.Error(); got != "RESOLVE_CREDENTIALS is set but SUPABASE_DB_URL is missing" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestResolveCredentials_MissingEncryptionKey(t *testing.T) {
	t.Setenv("RESOLVE_CREDENTIALS", "1")
	t.Setenv("SUPABASE_DB_URL", "postgres://localhost/test")
	os.Unsetenv("ENCRYPTION_MASTER_KEY")

	cmd := &cobra.Command{}
	err := resolveCredentials(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing ENCRYPTION_MASTER_KEY")
	}
	if got := err.Error(); got != "RESOLVE_CREDENTIALS is set but ENCRYPTION_MASTER_KEY is missing" {
		t.Fatalf("unexpected error: %s", got)
	}
}
