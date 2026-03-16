package gmail

import (
	"testing"
)

// ---- Provider ----

func TestProviderNew(t *testing.T) {
	p := New()
	if p.Name() != "gmail" {
		t.Errorf("expected name=gmail, got %s", p.Name())
	}
	if p.ServiceFactory == nil {
		t.Error("expected ServiceFactory to be set")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{ServiceFactory: newTestServiceFactory(server)}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Verify the gmail subcommand is registered.
	gmailCmd, _, err := root.Find([]string{"gmail"})
	if err != nil || gmailCmd == nil {
		t.Fatal("expected gmail command to be registered")
	}

	// Verify the messages subcommand is registered.
	messagesCmd, _, err := root.Find([]string{"gmail", "messages"})
	if err != nil || messagesCmd == nil {
		t.Fatal("expected gmail messages command to be registered")
	}

	subCmds := map[string]bool{}
	for _, c := range messagesCmd.Commands() {
		subCmds[c.Use] = true
	}
	for _, expected := range []string{"list", "get", "send"} {
		if !subCmds[expected] {
			t.Errorf("expected subcommand %q to be registered under messages", expected)
		}
	}
}
