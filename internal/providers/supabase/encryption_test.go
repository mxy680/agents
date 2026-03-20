package supabase

import (
	"strings"
	"testing"
)

func TestEncryptionGetText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "encryption", "get", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "root_key")
	mustContain(t, out, "some-root-key-id")
}

func TestEncryptionGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "encryption", "get", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"root_key"`)
	mustContain(t, out, `"some-root-key-id"`)
}

func TestEncryptionUpdateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "encryption", "update", "--ref=test-ref",
			`--config={"root_key":"new-key"}`, "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "test-ref")
}

func TestEncryptionUpdateWithConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "encryption", "update", "--ref=test-ref",
			`--config={"root_key":"new-key"}`})
		root.Execute()
	})

	mustContain(t, out, "root_key")
}

func TestEncryptionUpdateMissingConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	root.SetArgs([]string{"supabase", "encryption", "update", "--ref=test-ref"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when neither --config nor --config-file provided")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}

func TestEncryptionAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "encrypt", "get", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "root_key")
}
