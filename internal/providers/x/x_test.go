package x

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderName(t *testing.T) {
	p := New()
	if p.Name() != "x" {
		t.Errorf("expected name=x, got %s", p.Name())
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

	xCmd, _, err := root.Find([]string{"x"})
	if err != nil {
		t.Fatalf("expected x command: %v", err)
	}
	if xCmd.Name() != "x" {
		t.Errorf("expected command name=x, got %s", xCmd.Name())
	}

	subcommands := map[string]bool{
		"posts": false,
		"users": false,
	}
	for _, sub := range xCmd.Commands() {
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

	xCmd, _, err := root.Find([]string{"twitter"})
	if err != nil {
		t.Fatalf("expected twitter alias to resolve: %v", err)
	}
	if xCmd.Name() != "x" {
		t.Errorf("expected command name=x via alias, got %s", xCmd.Name())
	}
}
