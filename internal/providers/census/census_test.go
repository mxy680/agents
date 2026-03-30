package census

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
	if got := p.Name(); got != "census" {
		t.Errorf("Name() = %q, want %q", got, "census")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: DefaultClientFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify census command was registered.
	censusCmd, _, err := root.Find([]string{"census"})
	if err != nil {
		t.Fatalf("census command not found: %v", err)
	}
	if censusCmd.Use != "census" {
		t.Errorf("census command Use = %q, want %q", censusCmd.Use, "census")
	}

	// Verify alias.
	aliasCmd, _, err := root.Find([]string{"acs"})
	if err != nil {
		t.Fatalf("acs alias not found: %v", err)
	}
	if aliasCmd != censusCmd {
		t.Error("acs alias does not point to census command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"tracts"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range censusCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under census", name)
		}
	}
}
