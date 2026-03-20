package supabase

import (
	"strings"
	"testing"
)

func TestSecretsListText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "secrets", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "MY_SECRET")
	mustContain(t, out, "DB_URL")
	// Values should be masked in table output
	if strings.Contains(out, "super-secret-value") {
		t.Error("expected secret value to be masked in table output, but it was shown in full")
	}
	mustContain(t, out, "****")
}

func TestSecretsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "secrets", "list", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"name"`)
	mustContain(t, out, `"MY_SECRET"`)
	mustContain(t, out, `"value"`)
}

func TestSecretsCreateText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "secrets", "create", "--ref=test-ref", "--name=NEW_SECRET", "--value=my-value"})
		root.Execute()
	})

	mustContain(t, out, "Created secret")
	mustContain(t, out, "NEW_SECRET")
}

func TestSecretsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "secrets", "create", "--ref=test-ref", "--name=NEW_SECRET", "--value=my-value", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "NEW_SECRET")
}

func TestSecretsDeleteMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "secrets", "delete", "--ref=test-ref", "--name=MY_SECRET"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(execErr.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", execErr)
	}
}

func TestSecretsDeleteSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "secrets", "delete", "--ref=test-ref", "--name=MY_SECRET", "--confirm"})
		root.Execute()
	})

	mustContain(t, out, "Deleted secret")
	mustContain(t, out, "MY_SECRET")
}

func TestSecretsAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "secret", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "MY_SECRET")
}
