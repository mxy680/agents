package gcp

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newServicesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("services",
		newServicesListCmd(factory),
		newServicesEnableCmd(factory),
		newServicesDisableCmd(factory),
	)
}

func TestServicesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "list", "--project", "my-project-1")

	mustContain(t, output, "iap.googleapis.com")
	mustContain(t, output, "iam.googleapis.com")
	mustContain(t, output, "ENABLED")
}

func TestServicesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "list", "--project", "my-project-1", "--json")

	var results []ServiceSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &results))
	assert.Len(t, results, 2)
	assert.Contains(t, results[0].Name, "iap.googleapis.com")
	assert.Equal(t, "ENABLED", results[0].State)
}

func TestServicesList_UsesClientProject(t *testing.T) {
	// No --project flag; should fall back to client's projectID ("my-project-1").
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "list")

	mustContain(t, output, "iap.googleapis.com")
}

func TestServicesEnable_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "enable",
		"--project", "my-project-1",
		"--service", "iap.googleapis.com",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "iap.googleapis.com")
}

func TestServicesEnable_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "enable",
		"--project", "my-project-1",
		"--service", "iap.googleapis.com")

	mustContain(t, output, "Enabled")
	mustContain(t, output, "iap.googleapis.com")
}

func TestServicesEnable_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "enable",
		"--project", "my-project-1",
		"--service", "iap.googleapis.com",
		"--json")

	var result map[string]string
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "enabled", result["status"])
	assert.Equal(t, "iap.googleapis.com", result["service"])
	assert.Equal(t, "my-project-1", result["project"])
}

func TestServicesDisable_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "disable",
		"--project", "my-project-1",
		"--service", "iap.googleapis.com",
		"--confirm")

	mustContain(t, output, "Disabled")
	mustContain(t, output, "iap.googleapis.com")
}

func TestServicesDisable_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	err := runCmdErr(t, root, "services", "disable",
		"--project", "my-project-1",
		"--service", "iap.googleapis.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestServicesDisable_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(factory))
	output := runCmd(t, root, "services", "disable",
		"--project", "my-project-1",
		"--service", "iap.googleapis.com",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "iap.googleapis.com")
}
