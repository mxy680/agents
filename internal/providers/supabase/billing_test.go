package supabase

import (
	"strings"
	"testing"
)

func TestBillingAddonsList(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "billing", "addons", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "compute_2x")
}

func TestBillingAddonsApplyDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "billing", "addons", "apply",
			"--ref=test-ref", "--addon=compute_2x", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "compute_2x")
}

func TestBillingAddonsApply(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "billing", "addons", "apply",
			"--ref=test-ref", "--addon=compute_2x"})
		root.Execute()
	})

	mustContain(t, out, "Applied addon")
	mustContain(t, out, "compute_2x")
}

func TestBillingAddonsRemoveMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	root.SetArgs([]string{"supabase", "billing", "addons", "remove",
		"--ref=test-ref", "--addon=compute_2x"})
	err := root.Execute()

	if err == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(err.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", err)
	}
}

func TestBillingAddonsRemoveSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "billing", "addons", "remove",
			"--ref=test-ref", "--addon=compute_2x", "--confirm"})
		root.Execute()
	})

	mustContain(t, out, "Removed addon")
	mustContain(t, out, "compute_2x")
}

func TestBillingAddonsRemoveDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "billing", "addons", "remove",
			"--ref=test-ref", "--addon=compute_2x", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
}

func TestBillingAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "bill", "addon", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "compute_2x")
}
