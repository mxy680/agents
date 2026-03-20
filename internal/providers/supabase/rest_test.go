package supabase

import (
	"strings"
	"testing"
)

func TestRestGetText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "rest", "get", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "db_schema")
	mustContain(t, out, "public")
}

func TestRestGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "rest", "get", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"db_schema"`)
	mustContain(t, out, `"public"`)
}

func TestRestUpdateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "rest", "update", "--ref=test-ref",
			`--config={"max_rows":500}`, "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "test-ref")
}

func TestRestUpdateWithConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "rest", "update", "--ref=test-ref",
			`--config={"max_rows":500}`})
		root.Execute()
	})

	mustContain(t, out, "db_schema")
}

func TestRestUpdateMissingConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	root.SetArgs([]string{"supabase", "rest", "update", "--ref=test-ref"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when neither --config nor --config-file provided")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected 'required' in error, got: %v", err)
	}
}
