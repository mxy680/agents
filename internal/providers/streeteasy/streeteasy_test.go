package streeteasy

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.ClientFactory == nil {
		t.Fatal("ClientFactory is nil")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "streeteasy" {
		t.Errorf("Name() = %q, want %q", got, "streeteasy")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify streeteasy command was registered.
	seCmd, _, err := root.Find([]string{"streeteasy"})
	if err != nil {
		t.Fatalf("streeteasy command not found: %v", err)
	}
	if seCmd.Use != "streeteasy" {
		t.Errorf("streeteasy command Use = %q, want %q", seCmd.Use, "streeteasy")
	}

	// Verify alias.
	aliasCmd, _, err := root.Find([]string{"se"})
	if err != nil {
		t.Fatalf("se alias not found: %v", err)
	}
	if aliasCmd != seCmd {
		t.Error("se alias does not point to streeteasy command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"listings"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range seCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under streeteasy", name)
		}
	}
}
