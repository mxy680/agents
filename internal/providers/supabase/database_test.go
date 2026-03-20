package supabase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newDBTestEnv(t *testing.T) (*Provider, func(args ...string) string) {
	t.Helper()
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)
	p := &Provider{ClientFactory: factory}

	run := func(args ...string) string {
		root := newTestRootCmd()
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs(args)
			root.Execute()
		})
		return out
	}
	return p, run
}

func TestDBMigrationsText(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "migrations", "--ref=test-ref")
	mustContain(t, out, "20260101000000")
	mustContain(t, out, "create_users_table")
	mustContain(t, out, "20260102000000")
	mustContain(t, out, "add_email_column")
}

func TestDBMigrationsJSON(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "migrations", "--ref=test-ref", "--json")
	mustContain(t, out, `"version"`)
	mustContain(t, out, `"20260101000000"`)
}

func TestDBTypesText(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "types", "--ref=test-ref")
	mustContain(t, out, "export type Database")
	mustContain(t, out, "users")
}

func TestDBSSLGetText(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "ssl-enforcement", "get", "--ref=test-ref")
	mustContain(t, out, "Enforced:")
	mustContain(t, out, "true")
}

func TestDBSSLGetJSON(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "ssl", "get", "--ref=test-ref", "--json")
	mustContain(t, out, `"enforced"`)
}

func TestDBSSLUpdateText(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "ssl-enforcement", "update", "--ref=test-ref", "--enabled=true")
	mustContain(t, out, "SSL enforcement updated")
	mustContain(t, out, "enforced=true")
}

func TestDBSSLUpdateDryRun(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "ssl", "update", "--ref=test-ref", "--enabled=false", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "test-ref")
}

func TestDBSSLAlias(t *testing.T) {
	_, run := newDBTestEnv(t)
	// "ssl" alias for "ssl-enforcement"
	out := run("supabase", "db", "ssl", "get", "--ref=test-ref")
	mustContain(t, out, "Enforced:")
}

func TestDBJITGetText(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "jit-access", "get", "--ref=test-ref")
	mustContain(t, out, "enabled")
}

func TestDBJITGetJSON(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "jit", "get", "--ref=test-ref", "--json")
	mustContain(t, out, `"enabled"`)
}

func TestDBJITUpdateWithConfig(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "jit-access", "update", "--ref=test-ref", `--config={"allowed_roles":["service_role"]}`)
	mustContain(t, out, "enabled")
}

func TestDBJITUpdateWithConfigFile(t *testing.T) {
	_, run := newDBTestEnv(t)

	// Write a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "jit-config.json")
	if err := os.WriteFile(configPath, []byte(`{"allowed_roles":["authenticated"]}`), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	out := run("supabase", "db", "jit-access", "update", "--ref=test-ref", "--config-file="+configPath)
	mustContain(t, out, "enabled")
}

func TestDBJITUpdateMissingConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)
	p := &Provider{ClientFactory: factory}

	root := newTestRootCmd()
	p.RegisterCommands(root)
	root.SetArgs([]string{"supabase", "db", "jit-access", "update", "--ref=test-ref"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when neither --config nor --config-file is provided, got nil")
	}
	if !strings.Contains(err.Error(), "--config or --config-file is required") {
		t.Errorf("expected '--config or --config-file is required' in error, got: %v", err)
	}
}

func TestDBJITUpdateDryRun(t *testing.T) {
	_, run := newDBTestEnv(t)
	out := run("supabase", "db", "jit-access", "update", "--ref=test-ref", "--config={}", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
}

func TestDBJITAlias(t *testing.T) {
	_, run := newDBTestEnv(t)
	// "jit" alias for "jit-access"
	out := run("supabase", "db", "jit", "get", "--ref=test-ref")
	mustContain(t, out, "enabled")
}
