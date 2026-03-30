package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newProjectsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("projects",
		newProjectsListCmd(factory),
		newProjectsGetCmd(factory),
		newProjectsCreateCmd(factory),
		newProjectsUpdateCmd(factory),
		newProjectsDeleteCmd(factory),
	)
}

func TestProjectsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "list")

	mustContain(t, output, "my-nextjs-app")
	mustContain(t, output, "my-vite-app")
}

func TestProjectsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "list", "--json")

	var results []ProjectSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(results))
	}
	if results[0].ID != "prj_abc123" {
		t.Errorf("expected first project ID=prj_abc123, got %s", results[0].ID)
	}
	if results[0].Name != "my-nextjs-app" {
		t.Errorf("expected first project Name=my-nextjs-app, got %s", results[0].Name)
	}
	if results[0].Framework != "nextjs" {
		t.Errorf("expected first project Framework=nextjs, got %s", results[0].Framework)
	}
	if results[1].ID != "prj_def456" {
		t.Errorf("expected second project ID=prj_def456, got %s", results[1].ID)
	}
}

func TestProjectsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "get", "--project", "my-nextjs-app")

	mustContain(t, output, "prj_abc123")
	mustContain(t, output, "my-nextjs-app")
	mustContain(t, output, "nextjs")
	mustContain(t, output, "npm run build")
}

func TestProjectsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "get", "--project", "my-nextjs-app", "--json")

	var detail ProjectDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "prj_abc123" {
		t.Errorf("expected ID=prj_abc123, got %s", detail.ID)
	}
	if detail.Name != "my-nextjs-app" {
		t.Errorf("expected Name=my-nextjs-app, got %s", detail.Name)
	}
	if detail.BuildCommand != "npm run build" {
		t.Errorf("expected BuildCommand=npm run build, got %s", detail.BuildCommand)
	}
	if detail.OutputDirectory != ".next" {
		t.Errorf("expected OutputDirectory=.next, got %s", detail.OutputDirectory)
	}
	if detail.AccountID != "acct_123" {
		t.Errorf("expected AccountID=acct_123, got %s", detail.AccountID)
	}
}

func TestProjectsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--name", "new-project", "--framework", "nextjs", "--json")

	var detail ProjectDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "prj_created1" {
		t.Errorf("expected ID=prj_created1, got %s", detail.ID)
	}
	if detail.Name != "new-project" {
		t.Errorf("expected Name=new-project, got %s", detail.Name)
	}
}

func TestProjectsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--name", "new-project", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "new-project")
}

func TestProjectsUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "update", "--project", "my-nextjs-app", "--build-command", "npm run build:prod", "--json")

	var detail ProjectDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "prj_abc123" {
		t.Errorf("expected ID=prj_abc123, got %s", detail.ID)
	}
}

func TestProjectsDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "delete", "--project", "my-nextjs-app", "--confirm")

	mustContain(t, output, "Deleted")
	mustContain(t, output, "my-nextjs-app")
}

func TestProjectsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--name", "new-project")

	mustContain(t, output, "Created project")
	mustContain(t, output, "new-project")
}

func TestProjectsUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "update", "--project", "my-nextjs-app", "--name", "renamed-app")

	mustContain(t, output, "Updated project")
}

func TestProjectsUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "update", "--project", "my-nextjs-app", "--name", "renamed-app", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestProjectsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	err := runCmdErr(t, root, "projects", "delete", "--project", "my-nextjs-app")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}
