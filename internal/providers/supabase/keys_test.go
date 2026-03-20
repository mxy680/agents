package supabase

import (
	"strings"
	"testing"
)

func TestKeysListText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "keys", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "key-001")
	mustContain(t, out, "anon")
	// The full API key should be masked in table output
	if strings.Contains(out, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test-anon-key") {
		t.Error("expected API key to be masked in table output, but it was shown in full")
	}
	mustContain(t, out, "eyJh****")
}

func TestKeysListJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "keys", "list", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"id"`)
	mustContain(t, out, `"key-001"`)
	mustContain(t, out, `"apiKey"`)
}

func TestKeysGet(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "keys", "get", "--ref=test-ref", "--key-id=key-001"})
		root.Execute()
	})

	mustContain(t, out, "key-001")
	mustContain(t, out, "anon")
	// get shows the full API key
	mustContain(t, out, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test-anon-key")
}

func TestKeysCreateText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "keys", "create", "--ref=test-ref", "--name=my-key"})
		root.Execute()
	})

	mustContain(t, out, "Created API key")
}

func TestKeysCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "keys", "create", "--ref=test-ref", "--name=my-key", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "my-key")
}

func TestKeysUpdate(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "keys", "update", "--ref=test-ref", "--key-id=key-001", "--name=renamed-key"})
		root.Execute()
	})

	mustContain(t, out, "Updated API key")
}

func TestKeysDeleteMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "keys", "delete", "--ref=test-ref", "--key-id=key-001"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(execErr.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", execErr)
	}
}

func TestKeysDeleteSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "keys", "delete", "--ref=test-ref", "--key-id=key-001", "--confirm"})
		root.Execute()
	})

	mustContain(t, out, "Deleted API key")
	mustContain(t, out, "key-001")
}

func TestKeysAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	// Use the "key" alias instead of "keys"
	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "key", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "key-001")
}
