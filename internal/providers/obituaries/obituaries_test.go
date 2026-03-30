package obituaries

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
	if got := p.Name(); got != "obituaries" {
		t.Errorf("Name() = %q, want %q", got, "obituaries")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify obituaries command was registered.
	obitCmd, _, err := root.Find([]string{"obituaries"})
	if err != nil {
		t.Fatalf("obituaries command not found: %v", err)
	}
	if obitCmd.Use != "obituaries" {
		t.Errorf("obituaries command Use = %q, want %q", obitCmd.Use, "obituaries")
	}

	// Verify alias.
	aliasCmd, _, err := root.Find([]string{"obit"})
	if err != nil {
		t.Fatalf("obit alias not found: %v", err)
	}
	if aliasCmd != obitCmd {
		t.Error("obit alias does not point to obituaries command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"search", "names"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range obitCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under obituaries", name)
		}
	}
}
