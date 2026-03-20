package supabase

import (
	"strings"
	"testing"
)

func newDomainsTestEnv(t *testing.T) (*Provider, func(args ...string) string) {
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

// --- custom hostname tests ---

func TestDomainsCustomGetText(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "get", "--ref=test-ref")
	mustContain(t, out, "app.example.com")
	mustContain(t, out, "Active")
}

func TestDomainsCustomGetJSON(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "get", "--ref=test-ref", "--json")
	mustContain(t, out, `"custom_hostname"`)
	mustContain(t, out, "app.example.com")
}

func TestDomainsCustomDeleteMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)
	p := &Provider{ClientFactory: factory}

	root := newTestRootCmd()
	p.RegisterCommands(root)
	root.SetArgs([]string{"supabase", "domains", "custom", "delete", "--ref=test-ref"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(err.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", err)
	}
}

func TestDomainsCustomDeleteSuccess(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "delete", "--ref=test-ref", "--confirm")
	mustContain(t, out, "Deleted custom hostname")
	mustContain(t, out, "test-ref")
}

func TestDomainsCustomDeleteDryRun(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "delete", "--ref=test-ref", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "test-ref")
}

func TestDomainsCustomInitialize(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "initialize", "--ref=test-ref", "--hostname=myapp.example.com")
	mustContain(t, out, "myapp.example.com")
}

func TestDomainsCustomInitializeDryRun(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "initialize", "--ref=test-ref", "--hostname=myapp.example.com", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "myapp.example.com")
}

func TestDomainsCustomVerify(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "verify", "--ref=test-ref")
	mustContain(t, out, "reverification")
	mustContain(t, out, "test-ref")
}

func TestDomainsCustomVerifyDryRun(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "verify", "--ref=test-ref", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
}

func TestDomainsCustomActivate(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "activate", "--ref=test-ref")
	mustContain(t, out, "Activated custom hostname")
	mustContain(t, out, "test-ref")
}

func TestDomainsCustomActivateDryRun(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "custom", "activate", "--ref=test-ref", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
}

// --- vanity subdomain tests ---

func TestDomainsVanityGetText(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "get", "--ref=test-ref")
	mustContain(t, out, "my-app")
}

func TestDomainsVanityGetJSON(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "get", "--ref=test-ref", "--json")
	mustContain(t, out, `"vanity_subdomain"`)
	mustContain(t, out, "my-app")
}

func TestDomainsVanityDeleteMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)
	p := &Provider{ClientFactory: factory}

	root := newTestRootCmd()
	p.RegisterCommands(root)
	root.SetArgs([]string{"supabase", "domains", "vanity", "delete", "--ref=test-ref"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(err.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", err)
	}
}

func TestDomainsVanityDeleteSuccess(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "delete", "--ref=test-ref", "--confirm")
	mustContain(t, out, "Deleted vanity subdomain")
	mustContain(t, out, "test-ref")
}

func TestDomainsVanityDeleteDryRun(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "delete", "--ref=test-ref", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
}

func TestDomainsVanityCheckAvailable(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "check", "--ref=test-ref", "--subdomain=my-new-app")
	mustContain(t, out, "my-new-app")
	mustContain(t, out, "Available:")
	mustContain(t, out, "true")
}

func TestDomainsVanityCheckUnavailable(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "check", "--ref=test-ref", "--subdomain=taken-subdomain")
	mustContain(t, out, "Available:")
	mustContain(t, out, "false")
}

func TestDomainsVanityCheckJSON(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "check", "--ref=test-ref", "--subdomain=my-new-app", "--json")
	mustContain(t, out, `"available"`)
}

func TestDomainsVanityActivate(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "activate", "--ref=test-ref", "--subdomain=my-new-app")
	mustContain(t, out, "my-new-app")
	mustContain(t, out, "test-ref")
}

func TestDomainsVanityActivateDryRun(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	out := run("supabase", "domains", "vanity", "activate", "--ref=test-ref", "--subdomain=my-new-app", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "my-new-app")
}

func TestDomainsAlias(t *testing.T) {
	_, run := newDomainsTestEnv(t)
	// "domain" alias for "domains"
	out := run("supabase", "domain", "custom", "get", "--ref=test-ref")
	mustContain(t, out, "app.example.com")
}
