package imessage

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
		t.Error("New() ClientFactory is nil")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "imessage" {
		t.Errorf("Name() = %q, want %q", got, "imessage")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// The imessage command should be registered on root.
	var found *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "imessage" {
			found = cmd
			break
		}
	}
	if found == nil {
		t.Fatal("imessage command not found on root")
	}

	// Verify the alias is registered.
	hasAlias := false
	for _, a := range found.Aliases {
		if a == "imsg" {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		t.Error("imessage command missing alias 'imsg'")
	}

	// Verify expected subcommands are registered.
	wantSubcmds := []string{
		"chats",
		"participants",
		"messages",
		"scheduled",
		"attachments",
		"handles",
		"contacts",
		"facetime",
		"findmy",
		"icloud",
		"server",
		"webhooks",
		"mac",
	}

	registeredSubcmds := make(map[string]bool)
	for _, sub := range found.Commands() {
		registeredSubcmds[sub.Use] = true
	}

	for _, want := range wantSubcmds {
		if !registeredSubcmds[want] {
			t.Errorf("expected subcommand %q not found under imessage", want)
		}
	}
}

func TestProviderRegisterCommandsWithCustomFactory(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{
		ClientFactory: newTestClientFactory(server),
	}

	if p.Name() != "imessage" {
		t.Errorf("Name() = %q, want %q", p.Name(), "imessage")
	}

	root := newTestRootCmd()
	p.RegisterCommands(root)

	if root.Commands() == nil {
		t.Error("RegisterCommands did not add any commands")
	}
}
