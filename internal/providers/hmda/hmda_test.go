package hmda

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
	if got := p.Name(); got != "hmda" {
		t.Errorf("Name() = %q, want %q", got, "hmda")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify hmda command was registered.
	hmdaCmd, _, err := root.Find([]string{"hmda"})
	if err != nil {
		t.Fatalf("hmda command not found: %v", err)
	}
	if hmdaCmd.Use != "hmda" {
		t.Errorf("hmda command Use = %q, want %q", hmdaCmd.Use, "hmda")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"loans"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range hmdaCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under hmda", name)
		}
	}
}
