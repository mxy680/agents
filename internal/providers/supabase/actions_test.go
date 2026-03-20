package supabase

import (
	"testing"
)

func TestActionsListText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "actions", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "run-abc123")
	mustContain(t, out, "completed")
}

func TestActionsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "actions", "list", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"id"`)
	mustContain(t, out, `"run-abc123"`)
}

func TestActionsGet(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "actions", "get", "--ref=test-ref", "--run-id=run-abc123"})
		root.Execute()
	})

	mustContain(t, out, "run-abc123")
	mustContain(t, out, "completed")
}

func TestActionsLogs(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "actions", "logs", "--ref=test-ref", "--run-id=run-abc123"})
		root.Execute()
	})

	mustContain(t, out, "Deploy started")
}

func TestActionsLogsJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "actions", "logs", "--ref=test-ref", "--run-id=run-abc123", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"message"`)
}

func TestActionsUpdateStatusDryRun(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "actions", "update-status",
			"--ref=test-ref", "--run-id=run-abc123", "--status=cancelled", "--dry-run"})
		root.Execute()
	})

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "cancelled")
}

func TestActionsUpdateStatus(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "actions", "update-status",
			"--ref=test-ref", "--run-id=run-abc123", "--status=cancelled"})
		root.Execute()
	})

	mustContain(t, out, "run-abc123")
	mustContain(t, out, "cancelled")
}

func TestActionsAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "action", "list", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "run-abc123")
}
