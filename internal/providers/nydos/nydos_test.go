package nydos

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
	if got := p.Name(); got != "nydos" {
		t.Errorf("Name() = %q, want %q", got, "nydos")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify nydos command was registered.
	dosCmd, _, err := root.Find([]string{"nydos"})
	if err != nil {
		t.Fatalf("nydos command not found: %v", err)
	}
	if dosCmd.Use != "nydos" {
		t.Errorf("nydos command Use = %q, want %q", dosCmd.Use, "nydos")
	}

	// Verify alias works.
	aliasCmd, _, err := root.Find([]string{"dos"})
	if err != nil {
		t.Fatalf("dos alias not found: %v", err)
	}
	if aliasCmd == nil {
		t.Fatal("dos alias returned nil command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"entities"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range dosCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under nydos", name)
		}
	}
}
