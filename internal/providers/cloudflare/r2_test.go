package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newR2TestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("r2",
		newR2ListCmd(factory),
		newR2CreateCmd(factory),
		newR2DeleteCmd(factory),
	)
}

func TestR2List_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "list")

	mustContain(t, output, "my-bucket")
	mustContain(t, output, "another-bucket")
}

func TestR2List_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "list", "--json")

	var buckets []R2BucketSummary
	err := json.Unmarshal([]byte(output), &buckets)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, buckets, 2)
	assert.Equal(t, "my-bucket", buckets[0].Name)
	assert.Equal(t, "2024-01-01T00:00:00Z", buckets[0].CreationDate)
	assert.Equal(t, "another-bucket", buckets[1].Name)
}

func TestR2Create_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "create", "--name", "new-bucket")

	mustContain(t, output, "Created R2 bucket")
	mustContain(t, output, "new-bucket")
}

func TestR2Create_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "create", "--name", "new-bucket", "--json")

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "created", result["status"])
	assert.Equal(t, "new-bucket", result["name"])
}

func TestR2Create_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "create", "--name", "new-bucket", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "new-bucket")
}

func TestR2Delete_Confirm_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "delete", "--name", "my-bucket", "--confirm")

	mustContain(t, output, "Deleted R2 bucket")
	mustContain(t, output, "my-bucket")
}

func TestR2Delete_Confirm_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "delete", "--name", "my-bucket", "--confirm", "--json")

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "my-bucket", result["name"])
}

func TestR2Delete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))
	output := runCmd(t, root, "r2", "delete", "--name", "my-bucket", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-bucket")
}

func TestR2Delete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newR2TestCmd(factory))

	err := runCmdErr(t, root, "r2", "delete", "--name", "my-bucket")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
