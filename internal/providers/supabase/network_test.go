package supabase

import (
	"strings"
	"testing"
)

func newNetTestEnv(t *testing.T) (*Provider, func(args ...string) string) {
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

func TestNetRestrictionsGetText(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "restrictions", "get", "--ref=test-ref")
	mustContain(t, out, "allowed_cidrs")
	mustContain(t, out, "0.0.0.0/0")
}

func TestNetRestrictionsGetJSON(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "restrictions", "get", "--ref=test-ref", "--json")
	mustContain(t, out, `"allowed_cidrs"`)
}

func TestNetRestrictionsUpdateWithConfig(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "restrictions", "update", "--ref=test-ref", `--config={"allowed_cidrs":["192.168.1.0/24"]}`)
	mustContain(t, out, "allowed_cidrs")
}

func TestNetRestrictionsUpdateDryRun(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "restrictions", "update", "--ref=test-ref", `--config={}`, "--dry-run")
	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "test-ref")
}

func TestNetRestrictionsUpdateMissingConfig(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)
	p := &Provider{ClientFactory: factory}

	root := newTestRootCmd()
	p.RegisterCommands(root)
	root.SetArgs([]string{"supabase", "network", "restrictions", "update", "--ref=test-ref"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when neither --config nor --config-file is provided, got nil")
	}
	if !strings.Contains(err.Error(), "--config or --config-file is required") {
		t.Errorf("expected '--config or --config-file is required' in error, got: %v", err)
	}
}

func TestNetRestrictionsApply(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "restrictions", "apply", "--ref=test-ref")
	mustContain(t, out, "test-ref")
}

func TestNetRestrictionsApplyDryRun(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "restrictions", "apply", "--ref=test-ref", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
}

func TestNetBansListText(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "bans", "list", "--ref=test-ref")
	mustContain(t, out, "1.2.3.4")
	mustContain(t, out, "5.6.7.8")
}

func TestNetBansListJSON(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "bans", "list", "--ref=test-ref", "--json")
	mustContain(t, out, `"banned_ipv4_addresses"`)
	mustContain(t, out, "1.2.3.4")
}

func TestNetBansRemoveMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)
	p := &Provider{ClientFactory: factory}

	root := newTestRootCmd()
	p.RegisterCommands(root)
	root.SetArgs([]string{"supabase", "network", "bans", "remove", "--ref=test-ref", "--ips=1.2.3.4"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(err.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", err)
	}
}

func TestNetBansRemoveSuccess(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "bans", "remove", "--ref=test-ref", "--ips=1.2.3.4,5.6.7.8", "--confirm")
	mustContain(t, out, "1.2.3.4")
}

func TestNetBansRemoveDryRun(t *testing.T) {
	_, run := newNetTestEnv(t)
	out := run("supabase", "network", "bans", "remove", "--ref=test-ref", "--ips=1.2.3.4", "--dry-run")
	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "1.2.3.4")
}

func TestNetworkAlias(t *testing.T) {
	_, run := newNetTestEnv(t)
	// "net" alias for "network", "restrict" alias for "restrictions"
	out := run("supabase", "net", "restrict", "get", "--ref=test-ref")
	mustContain(t, out, "allowed_cidrs")
}

func TestNetworkBanAlias(t *testing.T) {
	_, run := newNetTestEnv(t)
	// "ban" alias for "bans"
	out := run("supabase", "network", "ban", "list", "--ref=test-ref")
	mustContain(t, out, "1.2.3.4")
}
