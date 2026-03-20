package linkedin

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderName(t *testing.T) {
	p := New()
	if p.Name() != "linkedin" {
		t.Errorf("expected name=linkedin, got %s", p.Name())
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

	liCmd, _, err := root.Find([]string{"linkedin"})
	if err != nil {
		t.Fatalf("expected linkedin command: %v", err)
	}
	if liCmd.Name() != "linkedin" {
		t.Errorf("expected command name=linkedin, got %s", liCmd.Name())
	}

	subcommands := map[string]bool{
		"profile":     false,
		"connections": false,
		"invitations": false,
		"posts":       false,
		"comments":    false,
		"feed":        false,
	}
	for _, sub := range liCmd.Commands() {
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

	liCmd, _, err := root.Find([]string{"li"})
	if err != nil {
		t.Fatalf("expected li alias to resolve: %v", err)
	}
	if liCmd.Name() != "linkedin" {
		t.Errorf("expected command name=linkedin via alias, got %s", liCmd.Name())
	}
}
