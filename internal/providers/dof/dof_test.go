package dof

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
	if got := p.Name(); got != "dof" {
		t.Errorf("Name() = %q, want %q", got, "dof")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify dof command was registered.
	dofCmd, _, err := root.Find([]string{"dof"})
	if err != nil {
		t.Fatalf("dof command not found: %v", err)
	}
	if dofCmd.Use != "dof" {
		t.Errorf("dof command Use = %q, want %q", dofCmd.Use, "dof")
	}

	// Verify alias "tax" is present.
	found := false
	for _, alias := range dofCmd.Aliases {
		if alias == "tax" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected alias 'tax' on dof command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"owners"}
	for _, name := range expectedSubs {
		subFound := false
		for _, sub := range dofCmd.Commands() {
			if sub.Name() == name {
				subFound = true
				break
			}
		}
		if !subFound {
			t.Errorf("subcommand %q not found under dof", name)
		}
	}
}
