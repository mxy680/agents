package supabase

import (
	"testing"
)

func TestSnippetsListText(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "snippets", "list"})
		root.Execute()
	})

	mustContain(t, out, "snippet-001")
	mustContain(t, out, "Get recent users")
}

func TestSnippetsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "snippets", "list", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"id"`)
	mustContain(t, out, `"snippet-001"`)
	mustContain(t, out, `"name"`)
}

func TestSnippetsGet(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "snippets", "get", "--snippet-id=snippet-001"})
		root.Execute()
	})

	mustContain(t, out, "snippet-001")
	mustContain(t, out, "Get recent users")
	mustContain(t, out, "Returns users created in the last 30 days")
}

func TestSnippetsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "snippets", "get", "--snippet-id=snippet-001", "--json"})
		root.Execute()
	})

	mustContain(t, out, `"id"`)
	mustContain(t, out, `"snippet-001"`)
}

func TestSnippetsAlias(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	p := &Provider{ClientFactory: factory}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"supabase", "snippet", "list"})
		root.Execute()
	})

	mustContain(t, out, "snippet-001")
}
