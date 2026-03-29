package gcp

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newIAMTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("iam",
		newIAMListCmd(factory),
		newIAMCreateCmd(factory),
		newIAMCreateKeyCmd(factory),
		newIAMDeleteCmd(factory),
	)
}

func TestIAMList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "list", "--project", "my-project-1")

	mustContain(t, output, "svc-acct-1@my-project-1.iam.gserviceaccount.com")
	mustContain(t, output, "svc-acct-2@my-project-1.iam.gserviceaccount.com")
}

func TestIAMList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "list", "--project", "my-project-1", "--json")

	var results []ServiceAccountSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &results))
	assert.Len(t, results, 2)
	assert.Equal(t, "svc-acct-1@my-project-1.iam.gserviceaccount.com", results[0].Email)
	assert.Equal(t, "Service Account One", results[0].DisplayName)
	assert.False(t, results[0].Disabled)
}

func TestIAMList_UsesClientProject(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "list")

	mustContain(t, output, "svc-acct-1@my-project-1.iam.gserviceaccount.com")
}

func TestIAMCreate_DryRun_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create",
		"--project", "my-project-1",
		"--account-id", "new-svc-acct",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "new-svc-acct")
}

func TestIAMCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create",
		"--project", "my-project-1",
		"--account-id", "new-svc-acct",
		"--dry-run", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "create", result["action"])
	assert.Equal(t, "new-svc-acct", result["accountId"])
}

func TestIAMCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create",
		"--project", "my-project-1",
		"--account-id", "svc-acct-1",
		"--display-name", "Service Account One")

	mustContain(t, output, "Created service account")
	mustContain(t, output, "svc-acct-1@my-project-1.iam.gserviceaccount.com")
}

func TestIAMCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create",
		"--project", "my-project-1",
		"--account-id", "svc-acct-1",
		"--display-name", "Service Account One",
		"--json")

	var result ServiceAccountSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "svc-acct-1@my-project-1.iam.gserviceaccount.com", result.Email)
	assert.Equal(t, "Service Account One", result.DisplayName)
}

func TestIAMCreateKey_DryRun_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create-key",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "svc-acct-1@my-project-1.iam.gserviceaccount.com")
}

func TestIAMCreateKey_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create-key",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com",
		"--dry-run", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "create-key", result["action"])
}

func TestIAMCreateKey_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create-key",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com")

	mustContain(t, output, "Private Key Data")
	mustContain(t, output, "base64-encoded-json-key-data")
	mustContain(t, output, "USER_MANAGED")
}

func TestIAMCreateKey_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "create-key",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com",
		"--json")

	var result ServiceAccountKey
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Contains(t, result.Name, "key123")
	assert.Equal(t, "base64-encoded-json-key-data", result.PrivateKeyData)
	assert.Equal(t, "USER_MANAGED", result.KeyType)
}

func TestIAMDelete_DryRun_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "delete",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "svc-acct-1@my-project-1.iam.gserviceaccount.com")
}

func TestIAMDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "delete",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com",
		"--confirm")

	mustContain(t, output, "Deleted service account")
	mustContain(t, output, "svc-acct-1@my-project-1.iam.gserviceaccount.com")
}

func TestIAMDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	err := runCmdErr(t, root, "iam", "delete",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestIAMDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIAMTestCmd(factory))
	output := runCmd(t, root, "iam", "delete",
		"--project", "my-project-1",
		"--email", "svc-acct-1@my-project-1.iam.gserviceaccount.com",
		"--confirm", "--json")

	var result map[string]string
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "svc-acct-1@my-project-1.iam.gserviceaccount.com", result["email"])
}
