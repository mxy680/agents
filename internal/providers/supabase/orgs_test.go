package supabase

import (
	"testing"
)

func TestOrgsList(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "orgs", "list"})
			_ = root.Execute()
		})
		mustContain(t, output, "Acme Corp")
		mustContain(t, output, "My Personal Org")
		mustContain(t, output, "org-uuid-1234")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "orgs", "list", "--json"})
			_ = root.Execute()
		})
		mustContain(t, output, `"id"`)
		mustContain(t, output, `"name"`)
		mustContain(t, output, "Acme Corp")
	})
}

func TestOrgsCreate(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "orgs", "create", "--name=New Org"})
			_ = root.Execute()
		})
		mustContain(t, output, "Created:")
		mustContain(t, output, "New Org")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "orgs", "create", "--name=New Org", "--json"})
			_ = root.Execute()
		})
		mustContain(t, output, `"id"`)
		mustContain(t, output, "New Org")
	})

	t.Run("dry-run", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "orgs", "create", "--name=New Org", "--dry-run"})
			_ = root.Execute()
		})
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "New Org")
	})
}
