package supabase

import (
	"testing"
)

func TestAdvisorsPerformanceText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "advisors", "performance", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "Missing index on foreign key")
	mustContain(t, out, "WARN")
}

func TestAdvisorsPerformanceJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "advisors", "performance", "--ref=test-ref", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"title"`)
	mustContain(t, out, `"severity"`)
	mustContain(t, out, "WARN")
}

func TestAdvisorsSecurityText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "advisors", "security", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "Missing index on foreign key")
}

func TestAdvisorsAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	// Use the "advisor" alias
	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "advisor", "performance", "--ref=test-ref"})
		root.Execute()
	})

	mustContain(t, out, "WARN")
}
