package citibike

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
	if got := p.Name(); got != "citibike" {
		t.Errorf("Name() = %q, want %q", got, "citibike")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify citibike command was registered.
	cbCmd, _, err := root.Find([]string{"citibike"})
	if err != nil {
		t.Fatalf("citibike command not found: %v", err)
	}
	if cbCmd.Use != "citibike" {
		t.Errorf("citibike command Use = %q, want %q", cbCmd.Use, "citibike")
	}

	// Verify alias.
	aliasCmd, _, err := root.Find([]string{"cb"})
	if err != nil {
		t.Fatalf("cb alias not found: %v", err)
	}
	if aliasCmd != cbCmd {
		t.Error("cb alias does not point to citibike command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"stations"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range cbCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under citibike", name)
		}
	}
}
