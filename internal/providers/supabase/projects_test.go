package supabase

import (
	"testing"
)

func TestProjectsList(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "list"})
			_ = root.Execute()
		})
		mustContain(t, output, "my-app")
		mustContain(t, output, "my-staging")
		mustContain(t, output, "us-east-1")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "list", "--json"})
			_ = root.Execute()
		})
		mustContain(t, output, `"id"`)
		mustContain(t, output, `"name"`)
		mustContain(t, output, "my-app")
	})
}

func TestProjectsGet(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "get", "--ref=test-ref"})
			_ = root.Execute()
		})
		mustContain(t, output, "my-app")
		mustContain(t, output, "test-ref")
		mustContain(t, output, "us-east-1")
		mustContain(t, output, "db.test-ref.supabase.co")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "get", "--ref=test-ref", "--json"})
			_ = root.Execute()
		})
		mustContain(t, output, `"id"`)
		mustContain(t, output, "test-ref")
		mustContain(t, output, `"databaseHost"`)
	})
}

func TestProjectsCreate(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "create",
				"--name=new-project", "--org-id=org-uuid-1234", "--region=us-east-1"})
			_ = root.Execute()
		})
		mustContain(t, output, "Created:")
		mustContain(t, output, "new-project")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "create",
				"--name=new-project", "--org-id=org-uuid-1234", "--region=us-east-1", "--json"})
			_ = root.Execute()
		})
		mustContain(t, output, `"id"`)
		mustContain(t, output, "new-project")
	})

	t.Run("dry-run", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "create",
				"--name=new-project", "--org-id=org-uuid-1234", "--region=us-east-1", "--dry-run"})
			_ = root.Execute()
		})
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "new-project")
	})
}

func TestProjectsUpdate(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "update",
				"--ref=test-ref", "--name=updated-project"})
			_ = root.Execute()
		})
		mustContain(t, output, "Updated:")
		mustContain(t, output, "updated-project")
	})

	t.Run("dry-run", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "update",
				"--ref=test-ref", "--dry-run"})
			_ = root.Execute()
		})
		mustContain(t, output, "[DRY RUN]")
		mustContain(t, output, "test-ref")
	})
}

func TestProjectsDelete(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("missing --confirm returns error", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		err := func() error {
			root.SetArgs([]string{"supabase", "projects", "delete", "--ref=test-ref"})
			return root.Execute()
		}()
		if err == nil {
			t.Fatal("expected error when --confirm is not provided")
		}
	})

	t.Run("success with --confirm", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "delete", "--ref=test-ref", "--confirm"})
			_ = root.Execute()
		})
		mustContain(t, output, "Deleted:")
		mustContain(t, output, "test-ref")
	})

	t.Run("dry-run", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "delete", "--ref=test-ref", "--dry-run"})
			_ = root.Execute()
		})
		mustContain(t, output, "[DRY RUN]")
	})
}

func TestProjectsPause(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "pause", "--ref=test-ref"})
			_ = root.Execute()
		})
		mustContain(t, output, "Paused:")
		mustContain(t, output, "test-ref")
	})

	t.Run("dry-run", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "pause", "--ref=test-ref", "--dry-run"})
			_ = root.Execute()
		})
		mustContain(t, output, "[DRY RUN]")
	})
}

func TestProjectsRestore(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "restore", "--ref=test-ref"})
			_ = root.Execute()
		})
		mustContain(t, output, "Restored:")
		mustContain(t, output, "test-ref")
	})

	t.Run("dry-run", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "restore", "--ref=test-ref", "--dry-run"})
			_ = root.Execute()
		})
		mustContain(t, output, "[DRY RUN]")
	})
}

func TestProjectsHealth(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "health", "--ref=test-ref"})
			_ = root.Execute()
		})
		mustContain(t, output, "database")
		mustContain(t, output, "HEALTHY")
		mustContain(t, output, "auth")
		mustContain(t, output, "storage")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "health", "--ref=test-ref", "--json"})
			_ = root.Execute()
		})
		mustContain(t, output, `"name"`)
		mustContain(t, output, `"status"`)
		mustContain(t, output, "database")
	})
}

func TestProjectsRegions(t *testing.T) {
	server := newFullMockServer(t)
	setEnv(t, "SUPABASE_API_BASE_URL", server.URL)
	factory := newTestClientFactory(server)

	t.Run("text output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "regions", "--org-slug=test-org"})
			_ = root.Execute()
		})
		mustContain(t, output, "us-east-1")
		mustContain(t, output, "us-west-2")
		mustContain(t, output, "eu-central-1")
	})

	t.Run("json output", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: factory}
		p.RegisterCommands(root)

		output := captureStdout(t, func() {
			root.SetArgs([]string{"supabase", "projects", "regions", "--org-slug=test-org", "--json"})
			_ = root.Execute()
		})
		mustContain(t, output, "us-east-1")
	})
}
