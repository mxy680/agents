package calendar

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderName(t *testing.T) {
	p := New()
	if p.Name() != "calendar" {
		t.Errorf("expected name=calendar, got %s", p.Name())
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

	calCmd, _, err := root.Find([]string{"calendar"})
	if err != nil {
		t.Fatalf("expected calendar command: %v", err)
	}
	if calCmd.Name() != "calendar" {
		t.Errorf("expected command name=calendar, got %s", calCmd.Name())
	}

	// Check subcommands exist
	subcommands := map[string]bool{
		"events":   false,
		"calendars": false,
		"freebusy": false,
	}
	for _, sub := range calCmd.Commands() {
		subcommands[sub.Name()] = true
	}
	for name, found := range subcommands {
		if !found {
			t.Errorf("expected subcommand %s not found", name)
		}
	}
}
