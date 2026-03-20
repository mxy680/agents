package supabase

import (
	"testing"
)

func TestAnalyticsLogs(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "analytics", "logs", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "result")
}

func TestAnalyticsAPICounts(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "analytics", "api-counts", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "result")
}

func TestAnalyticsAPIRequests(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "analytics", "api-requests", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "result")
}

func TestAnalyticsFunctions(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "analytics", "functions", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "result")
}

func TestAnalyticsLogsAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "logs", "logs", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "result")
}

func TestAnalyticsLogsJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "analytics", "logs", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"result"`)
}
