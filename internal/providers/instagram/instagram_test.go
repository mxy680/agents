package instagram

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderName(t *testing.T) {
	p := New()
	if p.Name() != "instagram" {
		t.Errorf("expected name=instagram, got %s", p.Name())
	}
}

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.ClientFactory == nil {
		t.Error("expected non-nil ClientFactory")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "test"}
	p.RegisterCommands(root)

	igCmd, _, err := root.Find([]string{"instagram"})
	if err != nil {
		t.Fatalf("expected instagram command: %v", err)
	}
	if igCmd.Name() != "instagram" {
		t.Errorf("expected command name=instagram, got %s", igCmd.Name())
	}

	// Check that the 'profile' subcommand exists
	subcommands := map[string]bool{
		"profile": false,
	}
	for _, sub := range igCmd.Commands() {
		subcommands[sub.Name()] = true
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("expected subcommand %s not found", name)
		}
	}
}

func TestProviderRegisterCommandsAlias(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "test"}
	p.RegisterCommands(root)

	// Alias 'ig' should also resolve
	igCmd, _, err := root.Find([]string{"ig"})
	if err != nil {
		t.Fatalf("expected ig alias to resolve: %v", err)
	}
	if igCmd.Name() != "instagram" {
		t.Errorf("expected command name=instagram via alias, got %s", igCmd.Name())
	}
}
