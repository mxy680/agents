package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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

	mustContain(t, output, "My Project")
	mustContain(t, output, "Another Project")
	mustContain(t, output, "started")
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
	assert.Len(t, results, 2)
	assert.Equal(t, "proj-abc1", results[0].ID)
	assert.Equal(t, "My Project", results[0].Name)
	assert.Equal(t, "started", results[0].State)
	assert.InDelta(t, 0.4, results[0].Progress, 0.001)
	assert.Contains(t, results[0].Teams, "Engineering")
	assert.Equal(t, "proj-def2", results[1].ID)
	assert.Equal(t, "Another Project", results[1].Name)
}

func TestProjectsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "get", "--id", "proj-abc1")

	mustContain(t, output, "proj-abc1")
	mustContain(t, output, "My Project")
	mustContain(t, output, "started")
	mustContain(t, output, "A great project")
	mustContain(t, output, "2024-01-01")
	mustContain(t, output, "2024-06-01")
}

func TestProjectsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "get", "--id", "proj-abc1", "--json")

	var detail ProjectDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "proj-abc1", detail.ID)
	assert.Equal(t, "My Project", detail.Name)
	assert.Equal(t, "A great project", detail.Description)
	assert.Equal(t, "started", detail.State)
	assert.Equal(t, "2024-01-01", detail.StartDate)
	assert.Equal(t, "2024-06-01", detail.TargetDate)
	assert.InDelta(t, 0.4, detail.Progress, 0.001)
}

func TestProjectsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--name", "New Project", "--team", "team-abc1")

	mustContain(t, output, "Created project")
	mustContain(t, output, "New Project")
}

func TestProjectsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--name", "New Project", "--team", "team-abc1", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "proj-new1", result["id"])
	assert.Equal(t, "New Project", result["name"])
}

func TestProjectsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--name", "New Project", "--team", "team-abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "New Project")
}

func TestProjectsUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "update", "--id", "proj-abc1", "--name", "Renamed Project")

	mustContain(t, output, "Updated project")
}

func TestProjectsUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "update", "--id", "proj-abc1", "--name", "Renamed Project", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "proj-abc1", result["id"])
}

func TestProjectsUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "update", "--id", "proj-abc1", "--name", "Renamed Project", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestProjectsDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "delete", "--id", "proj-abc1", "--confirm")

	mustContain(t, output, "Deleted project")
	mustContain(t, output, "proj-abc1")
}

func TestProjectsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	err := runCmdErr(t, root, "projects", "delete", "--id", "proj-abc1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestProjectsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "delete", "--id", "proj-abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestProjectsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "delete", "--id", "proj-abc1", "--confirm", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "proj-abc1", result["id"])
}
