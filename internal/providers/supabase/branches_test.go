package supabase

import (
	"strings"
	"testing"
)

func TestBranchesListText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "branch-001")
	mustContain(t, out, "feat-login")
	mustContain(t, out, "ACTIVE_HEALTHY")
}

func TestBranchesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "list", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"id"`)
	mustContain(t, out, `"branch-001"`)
	mustContain(t, out, `"status"`)
}

func TestBranchesGet(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "get", "--branch-id=branch-001"})
		root.Execute()
	})

	mustContain(t, out, "branch-001")
	mustContain(t, out, "feat-login")
	mustContain(t, out, "feat/login")
	mustContain(t, out, "ACTIVE_HEALTHY")
}

func TestBranchesCreateText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "create", "--ref=test-ref", "--git-branch=feat/login"})
		root.Execute()
	})

	mustContain(t, out, "Created branch")
}

func TestBranchesCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "create", "--ref=test-ref", "--git-branch=feat/login", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "feat/login")
}

func TestBranchesUpdate(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "update", "--branch-id=branch-001", "--git-branch=feat/new"})
		root.Execute()
	})

	mustContain(t, out, "Updated branch")
}

func TestBranchesDeleteMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "branches", "delete", "--branch-id=branch-001"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(execErr.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", execErr)
	}
}

func TestBranchesDeleteSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "delete", "--branch-id=branch-001", "--confirm"})
		root.Execute()
	})

	mustContain(t, out, "Deleted branch")
	mustContain(t, out, "branch-001")
}

func TestBranchesPushText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "push", "--branch-id=branch-001"})
		root.Execute()
	})

	mustContain(t, out, "Push initiated")
	mustContain(t, out, "branch-001")
}

func TestBranchesPushDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "push", "--branch-id=branch-001", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "branch-001")
}

func TestBranchesMergeText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "merge", "--branch-id=branch-001"})
		root.Execute()
	})

	mustContain(t, out, "Merge initiated")
	mustContain(t, out, "branch-001")
}

func TestBranchesMergeDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "merge", "--branch-id=branch-001", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
}

func TestBranchesReset(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "reset", "--branch-id=branch-001"})
		root.Execute()
	})

	mustContain(t, out, "Reset initiated")
	mustContain(t, out, "branch-001")
}

func TestBranchesDiff(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "diff", "--branch-id=branch-001"})
		root.Execute()
	})

	mustContain(t, out, "diff")
}

func TestBranchesDisableMissingConfirm(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	var execErr error
	root.SetArgs([]string{"supabase", "branches", "disable", "--ref=test-ref"})
	execErr = root.Execute()

	if execErr == nil {
		t.Error("expected error when --confirm is missing, got nil")
	}
	if !strings.Contains(execErr.Error(), "re-run with --confirm") {
		t.Errorf("expected 're-run with --confirm' in error, got: %v", execErr)
	}
}

func TestBranchesDisableSuccess(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branches", "disable", "--ref=test-ref", "--confirm"})
		root.Execute()
	})

	mustContain(t, out, "Branching disabled")
	mustContain(t, out, "test-ref")
}

func TestBranchesAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	// Use the "branch" alias instead of "branches"
	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "branch", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "branch-001")
}
