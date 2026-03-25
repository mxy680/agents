package nysla

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
	if got := p.Name(); got != "nysla" {
		t.Errorf("Name() = %q, want %q", got, "nysla")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify nysla command was registered.
	nyslaCmd, _, err := root.Find([]string{"nysla"})
	if err != nil {
		t.Fatalf("nysla command not found: %v", err)
	}
	if nyslaCmd.Use != "nysla" {
		t.Errorf("nysla command Use = %q, want %q", nyslaCmd.Use, "nysla")
	}

	// Verify alias.
	aliasCmd, _, err := root.Find([]string{"liquor"})
	if err != nil {
		t.Fatalf("liquor alias not found: %v", err)
	}
	if aliasCmd != nyslaCmd {
		t.Error("liquor alias does not point to nysla command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"licenses"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range nyslaCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under nysla", name)
		}
	}
}
