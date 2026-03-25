package nyscef

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
	if got := p.Name(); got != "nyscef" {
		t.Errorf("Name() = %q, want %q", got, "nyscef")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{
		ClientFactory: newTestClientFactory(server),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify nyscef command was registered.
	nyscefCmd, _, err := root.Find([]string{"nyscef"})
	if err != nil {
		t.Fatalf("nyscef command not found: %v", err)
	}
	if nyscefCmd.Use != "nyscef" {
		t.Errorf("nyscef command Use = %q, want %q", nyscefCmd.Use, "nyscef")
	}

	// Verify subcommand exists.
	expectedSubs := []string{"cases"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range nyscefCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under nyscef", name)
		}
	}
}

func TestProviderRegisterCommandsSubcommands(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{
		ClientFactory: newTestClientFactory(server),
	}

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify cases subcommands exist.
	nyscefCmd, _, _ := root.Find([]string{"nyscef"})
	casesCmd, _, _ := nyscefCmd.Find([]string{"cases"})
	if casesCmd == nil {
		t.Fatal("cases command not found")
	}

	expectedCasesSubs := []string{"search", "get"}
	for _, name := range expectedCasesSubs {
		found := false
		for _, sub := range casesCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under cases", name)
		}
	}
}
