package supabase

import (
	"strings"
	"testing"
)

// --- auth get ---

func TestAuthGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "get", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"site_url"`)
	mustContain(t, out, `"jwt_exp"`)
	mustContain(t, out, "localhost:3000")
}

func TestAuthGetText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "get", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "site_url")
	mustContain(t, out, "localhost:3000")
}

// --- auth update ---

func TestAuthUpdateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "update", "--ref=test-ref", "--config={}", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "auth config")
}

func TestAuthUpdateSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "update", "--ref=test-ref", "--config={}"})
		root.Execute()
	})

	mustContain(t, out, "site_url")
}

func TestAuthUpdateMissingConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "auth", "update", "--ref=test-ref"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when neither --config nor --config-file is provided, got nil")
	}
	if !strings.Contains(execErr.Error(), "--config") {
		t.Errorf("expected '--config' in error, got: %v", execErr)
	}
}

// --- signing-keys list ---

func TestSigningKeysListJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "signing-keys", "list", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"sk-001"`)
	mustContain(t, out, `"active"`)
}

func TestSigningKeysListText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "signing-keys", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "sk-001")
}

// --- signing-keys get ---

func TestSigningKeysGet(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "signing-keys", "get", "--ref=test-ref", "--key-id=sk-001"})
		root.Execute()
	})

	mustContain(t, out, "sk-001")
	mustContain(t, out, "active")
}

// --- signing-keys create ---

func TestSigningKeysCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "signing-keys", "create", "--ref=test-ref", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "signing key")
}

func TestSigningKeysCreateSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "signing-keys", "create", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "sk-001")
}

// --- signing-keys delete ---

func TestSigningKeysDeleteMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "auth", "signing-keys", "delete", "--ref=test-ref", "--key-id=sk-001"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(execErr.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", execErr)
	}
}

func TestSigningKeysDeleteSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "signing-keys", "delete", "--ref=test-ref", "--key-id=sk-001", "--confirm"})
		root.Execute()
	})

	mustContain(t, out, "Deleted signing key")
	mustContain(t, out, "sk-001")
}

// --- signing-keys alias ---

func TestSigningKeysAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "sk", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "sk-001")
}

// --- third-party list ---

func TestThirdPartyListJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "third-party", "list", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"tpa-001"`)
	mustContain(t, out, `"google"`)
}

func TestThirdPartyListText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "third-party", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "tpa-001")
}

// --- third-party get ---

func TestThirdPartyGet(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "third-party", "get", "--ref=test-ref", "--tpa-id=tpa-001"})
		root.Execute()
	})

	mustContain(t, out, "tpa-001")
	mustContain(t, out, "google")
}

// --- third-party create ---

func TestThirdPartyCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "third-party", "create", "--ref=test-ref", "--config={}", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "third-party auth provider")
}

func TestThirdPartyCreateSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "third-party", "create", "--ref=test-ref", "--config={}"})
		root.Execute()
	})

	mustContain(t, out, "tpa-001")
}

func TestThirdPartyCreateMissingConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "auth", "third-party", "create", "--ref=test-ref"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when neither --config nor --config-file is provided, got nil")
	}
	if !strings.Contains(execErr.Error(), "--config") {
		t.Errorf("expected '--config' in error, got: %v", execErr)
	}
}

// --- third-party delete ---

func TestThirdPartyDeleteMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "auth", "third-party", "delete", "--ref=test-ref", "--tpa-id=tpa-001"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(execErr.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", execErr)
	}
}

func TestThirdPartyDeleteSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "third-party", "delete", "--ref=test-ref", "--tpa-id=tpa-001", "--confirm"})
		root.Execute()
	})

	mustContain(t, out, "Deleted third-party auth provider")
	mustContain(t, out, "tpa-001")
}

// --- third-party alias ---

func TestThirdPartyAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "auth", "tpa", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "tpa-001")
}
