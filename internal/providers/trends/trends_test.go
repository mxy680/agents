package trends

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.ServiceFactory == nil {
		t.Fatal("ServiceFactory is nil")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "trends" {
		t.Errorf("Name() = %q, want %q", got, "trends")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ServiceFactory: DefaultServiceFactory(),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify trends command was registered.
	gtCmd, _, err := root.Find([]string{"trends"})
	if err != nil {
		t.Fatalf("trends command not found: %v", err)
	}
	if gtCmd.Use != "trends" {
		t.Errorf("trends command Use = %q, want %q", gtCmd.Use, "trends")
	}

	// Verify alias.
	aliasCmd, _, err := root.Find([]string{"gt"})
	if err != nil {
		t.Fatalf("gt alias not found: %v", err)
	}
	if aliasCmd != gtCmd {
		t.Error("gt alias does not point to trends command")
	}

	// Verify subcommands exist.
	expectedSubs := []string{"interest"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range gtCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under trends", name)
		}
	}
}
