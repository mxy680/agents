package docs

import (
	"testing"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p.Name() != "docs" {
		t.Errorf("expected name=docs, got %s", p.Name())
	}
	if p.DocsServiceFactory == nil {
		t.Error("expected DocsServiceFactory to be set")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{
		DocsServiceFactory: newTestDocsServiceFactory(server),
	}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Verify the docs subcommand is registered.
	docsCmd, _, err := root.Find([]string{"docs"})
	if err != nil || docsCmd == nil {
		t.Fatal("expected docs command to be registered")
	}

	// Verify the documents subcommand group.
	cmd, _, err := root.Find([]string{"docs", "documents"})
	if err != nil || cmd == nil {
		t.Fatal("expected docs documents command to be registered")
	}

	// Verify documents subcommands.
	documentsCmd, _, _ := root.Find([]string{"docs", "documents"})
	subCmds := map[string]bool{}
	for _, c := range documentsCmd.Commands() {
		subCmds[c.Use] = true
	}
	for _, expected := range []string{"create", "get", "append", "batch-update"} {
		if !subCmds[expected] {
			t.Errorf("expected subcommand %q to be registered under documents", expected)
		}
	}
}
