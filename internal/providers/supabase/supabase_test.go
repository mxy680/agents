package supabase

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.ClientFactory == nil {
		t.Fatal("expected non-nil ClientFactory")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "supabase" {
		t.Errorf("Name() = %q, want %q", got, "supabase")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")

	p.RegisterCommands(root)

	// Verify supabase command was added
	supabaseCmd, _, err := root.Find([]string{"supabase"})
	if err != nil {
		t.Fatalf("supabase command not found: %v", err)
	}
	if supabaseCmd == nil {
		t.Fatal("expected non-nil supabase command")
	}
	if supabaseCmd.Name() != "supabase" {
		t.Errorf("command name = %q, want %q", supabaseCmd.Name(), "supabase")
	}
}

func TestProviderAliases(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")

	p.RegisterCommands(root)

	// Find the supabase command via alias "sb"
	sbCmd, _, err := root.Find([]string{"sb"})
	if err != nil {
		t.Fatalf("sb alias not found: %v", err)
	}
	if sbCmd == nil || sbCmd.Name() != "supabase" {
		t.Errorf("expected alias 'sb' to resolve to 'supabase' command, got %v", sbCmd)
	}
}
