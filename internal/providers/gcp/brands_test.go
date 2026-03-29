package gcp

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newBrandsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("brands",
		newBrandsListCmd(factory),
		newBrandsCreateCmd(factory),
		newBrandsGetCmd(factory),
	)
}

func TestBrandsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "list", "--project", "my-project-1")

	mustContain(t, output, "My Application")
	mustContain(t, output, "support@example.com")
}

func TestBrandsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "list", "--project", "my-project-1", "--json")

	var results []BrandSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &results))
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Name, "123456789")
	assert.Equal(t, "My Application", results[0].ApplicationTitle)
	assert.Equal(t, "support@example.com", results[0].SupportEmail)
	assert.True(t, results[0].OrgInternalOnly)
}

func TestBrandsList_UsesClientProject(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "list")

	mustContain(t, output, "My Application")
}

func TestBrandsCreate_DryRun_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "create",
		"--project", "my-project-1",
		"--title", "New App",
		"--support-email", "hello@example.com",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "New App")
}

func TestBrandsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "create",
		"--project", "my-project-1",
		"--title", "My Application",
		"--support-email", "support@example.com")

	mustContain(t, output, "Created brand")
	mustContain(t, output, "brands/123456789")
}

func TestBrandsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "create",
		"--project", "my-project-1",
		"--title", "My Application",
		"--support-email", "support@example.com",
		"--json")

	var result BrandSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Contains(t, result.Name, "brands/123456789")
	assert.Equal(t, "My Application", result.ApplicationTitle)
	assert.Equal(t, "support@example.com", result.SupportEmail)
}

func TestBrandsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "get",
		"--project", "my-project-1",
		"--brand", "123456789")

	mustContain(t, output, "My Application")
	mustContain(t, output, "support@example.com")
	mustContain(t, output, "true")
}

func TestBrandsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newBrandsTestCmd(factory))
	output := runCmd(t, root, "brands", "get",
		"--project", "my-project-1",
		"--brand", "123456789",
		"--json")

	var result BrandSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Contains(t, result.Name, "brands/123456789")
	assert.Equal(t, "My Application", result.ApplicationTitle)
	assert.Equal(t, "support@example.com", result.SupportEmail)
	assert.True(t, result.OrgInternalOnly)
}
