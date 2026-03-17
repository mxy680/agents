package drive

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderName(t *testing.T) {
	p := New()
	if p.Name() != "drive" {
		t.Errorf("expected name=drive, got %s", p.Name())
	}
}

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.ServiceFactory == nil {
		t.Error("expected non-nil ServiceFactory")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "test"}
	p.RegisterCommands(root)

	driveCmd, _, err := root.Find([]string{"drive"})
	if err != nil {
		t.Fatalf("expected drive command: %v", err)
	}
	if driveCmd.Name() != "drive" {
		t.Errorf("expected command name=drive, got %s", driveCmd.Name())
	}

	// Check subcommands exist
	subcommands := map[string]bool{
		"files":       false,
		"permissions": false,
	}
	for _, sub := range driveCmd.Commands() {
		subcommands[sub.Name()] = true
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("expected subcommand %s not found", name)
		}
	}
}
