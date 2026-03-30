package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveUserCredentials_NoOp(t *testing.T) {
	// When RESOLVE_USER_ID is not set, should be a no-op
	os.Unsetenv("RESOLVE_USER_ID")

	cmd := &cobra.Command{}
	if err := resolveUserCredentials(cmd, nil); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestResolveUserCredentials_MissingDBURL(t *testing.T) {
	t.Setenv("RESOLVE_USER_ID", "test-user-id")
	os.Unsetenv("SUPABASE_DB_URL")

	cmd := &cobra.Command{}
	err := resolveUserCredentials(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing SUPABASE_DB_URL")
	}
	if got := err.Error(); got != "RESOLVE_USER_ID is set but SUPABASE_DB_URL is missing" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestResolveUserCredentials_MissingEncryptionKey(t *testing.T) {
	t.Setenv("RESOLVE_USER_ID", "test-user-id")
	t.Setenv("SUPABASE_DB_URL", "postgres://localhost/test")
	os.Unsetenv("ENCRYPTION_MASTER_KEY")

	cmd := &cobra.Command{}
	err := resolveUserCredentials(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing ENCRYPTION_MASTER_KEY")
	}
	if got := err.Error(); got != "RESOLVE_USER_ID is set but ENCRYPTION_MASTER_KEY is missing" {
		t.Fatalf("unexpected error: %s", got)
	}
}
