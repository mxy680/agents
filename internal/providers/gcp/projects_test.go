package gcp

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

	mustContain(t, output, "my-project-1")
	mustContain(t, output, "my-project-2")
	mustContain(t, output, "ACTIVE")
}

func TestProjectsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "list", "--json")

	var results []ProjectSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &results))
	assert.Len(t, results, 2)
	assert.Equal(t, "my-project-1", results[0].ProjectID)
	assert.Equal(t, "My First Project", results[0].DisplayName)
	assert.Equal(t, "ACTIVE", results[0].State)
	assert.Equal(t, "my-project-2", results[1].ProjectID)
}

func TestProjectsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "get", "--project", "my-project-1")

	mustContain(t, output, "my-project-1")
	mustContain(t, output, "My First Project")
	mustContain(t, output, "ACTIVE")
	mustContain(t, output, "organizations/123456")
}

func TestProjectsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "get", "--project", "my-project-1", "--json")

	var detail ProjectDetail
	assert.NoError(t, json.Unmarshal([]byte(output), &detail))
	assert.Equal(t, "my-project-1", detail.ProjectID)
	assert.Equal(t, "My First Project", detail.DisplayName)
	assert.Equal(t, "ACTIVE", detail.State)
	assert.Equal(t, "organizations/123456", detail.Parent)
	assert.Equal(t, "etag-abc123", detail.Etag)
}

func TestProjectsCreate_DryRun_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--project", "new-proj", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "new-proj")
}

func TestProjectsCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--project", "new-proj", "--display-name", "New Project", "--dry-run", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "create", result["action"])
	assert.Equal(t, "new-proj", result["projectId"])
	assert.Equal(t, "New Project", result["displayName"])
}

func TestProjectsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--project", "new-proj")

	mustContain(t, output, "Created project")
	mustContain(t, output, "new-proj")
}

func TestProjectsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "create", "--project", "new-proj", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "created", result["status"])
	assert.Equal(t, "new-proj", result["projectId"])
}

func TestProjectsDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "delete", "--project", "my-project-1", "--confirm")

	mustContain(t, output, "Deleted project")
	mustContain(t, output, "my-project-1")
}

func TestProjectsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	err := runCmdErr(t, root, "projects", "delete", "--project", "my-project-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestProjectsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))
	output := runCmd(t, root, "projects", "delete", "--project", "my-project-1", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-project-1")
}
