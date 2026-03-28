package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newPagesTestCmd(factory ClientFactory) *cobra.Command {
	deploymentsCmd := buildTestCmd("deployments",
		newPagesDeploymentsListCmd(factory),
		newPagesDeploymentsGetCmd(factory),
	)
	return buildTestCmd("pages",
		newPagesListCmd(factory),
		newPagesGetCmd(factory),
		deploymentsCmd,
	)
}

func TestPagesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "list")

	mustContain(t, output, "my-pages-app")
	mustContain(t, output, "another-pages-app")
	mustContain(t, output, "main")
}

func TestPagesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "list", "--json")

	var projects []PagesSummary
	err := json.Unmarshal([]byte(output), &projects)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, projects, 2)
	assert.Equal(t, "pages_abc1", projects[0].ID)
	assert.Equal(t, "my-pages-app", projects[0].Name)
	assert.Equal(t, "my-pages-app.pages.dev", projects[0].SubDomain)
	assert.Equal(t, "main", projects[0].ProductionBranch)
	assert.Equal(t, "pages_def2", projects[1].ID)
}

func TestPagesGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "get", "--project", "my-pages-app")

	mustContain(t, output, "pages_abc1")
	mustContain(t, output, "my-pages-app")
	mustContain(t, output, "my-pages-app.pages.dev")
	mustContain(t, output, "main")
}

func TestPagesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "get", "--project", "my-pages-app", "--json")

	var project PagesSummary
	err := json.Unmarshal([]byte(output), &project)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "pages_abc1", project.ID)
	assert.Equal(t, "my-pages-app", project.Name)
	assert.Equal(t, "my-pages-app.pages.dev", project.SubDomain)
	assert.Equal(t, "main", project.ProductionBranch)
}

func TestPagesDeploymentsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "deployments", "list", "--project", "my-pages-app")

	mustContain(t, output, "deploy_abc1")
	mustContain(t, output, "production")
	mustContain(t, output, "preview")
}

func TestPagesDeploymentsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "deployments", "list", "--project", "my-pages-app", "--json")

	var deployments []PagesDeploymentSummary
	err := json.Unmarshal([]byte(output), &deployments)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, deployments, 2)
	assert.Equal(t, "deploy_abc1", deployments[0].ID)
	assert.Equal(t, "production", deployments[0].Environment)
	assert.Equal(t, "deploy", deployments[0].Stage)
	assert.Equal(t, "deploy_def2", deployments[1].ID)
}

func TestPagesDeploymentsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "deployments", "get",
		"--project", "my-pages-app",
		"--deployment", "deploy_abc1",
	)

	mustContain(t, output, "deploy_abc1")
	mustContain(t, output, "production")
	mustContain(t, output, "https://abc1.my-pages-app.pages.dev")
}

func TestPagesDeploymentsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newPagesTestCmd(factory))
	output := runCmd(t, root, "pages", "deployments", "get",
		"--project", "my-pages-app",
		"--deployment", "deploy_abc1",
		"--json",
	)

	var deployment PagesDeploymentSummary
	err := json.Unmarshal([]byte(output), &deployment)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "deploy_abc1", deployment.ID)
	assert.Equal(t, "production", deployment.Environment)
	assert.Equal(t, "deploy", deployment.Stage)
	assert.Equal(t, "https://abc1.my-pages-app.pages.dev", deployment.URL)
}
