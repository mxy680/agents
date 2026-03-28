package cloudflare

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newWorkersTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("workers",
		newWorkersListCmd(factory),
		newWorkersGetCmd(factory),
		newWorkersDeployCmd(factory),
		newWorkersDeleteCmd(factory),
	)
}

func TestWorkersList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "list")

	mustContain(t, output, "my-worker")
	mustContain(t, output, "another-worker")
}

func TestWorkersList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "list", "--json")

	var workers []WorkerSummary
	err := json.Unmarshal([]byte(output), &workers)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, workers, 2)
	assert.Equal(t, "my-worker", workers[0].ID)
	assert.Equal(t, "etag_abc1", workers[0].ETAG)
	assert.Equal(t, "2024-06-01T00:00:00Z", workers[0].ModifiedOn)
	assert.Equal(t, "another-worker", workers[1].ID)
}

func TestWorkersGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "get", "--name", "my-worker")

	mustContain(t, output, "my-worker")
	mustContain(t, output, "etag_abc1")
	mustContain(t, output, "2024-06-01T00:00:00Z")
}

func TestWorkersGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "get", "--name", "my-worker", "--json")

	var worker WorkerSummary
	err := json.Unmarshal([]byte(output), &worker)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "my-worker", worker.ID)
	assert.Equal(t, "etag_abc1", worker.ETAG)
}

func TestWorkersDeploy_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "deploy",
		"--name", "my-worker",
		"--file", "/tmp/worker.js",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-worker")
}

func TestWorkersDeploy_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	// Create a temp file with JS content
	tmpFile, err := os.CreateTemp(t.TempDir(), "worker-*.js")
	assert.NoError(t, err)
	_, err = tmpFile.WriteString("addEventListener('fetch', event => { event.respondWith(new Response('Hello')) })")
	assert.NoError(t, err)
	tmpFile.Close()

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "deploy",
		"--name", "my-worker",
		"--file", tmpFile.Name(),
	)

	mustContain(t, output, "Deployed worker")
	mustContain(t, output, "my-worker")
}

func TestWorkersDeploy_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	tmpFile, err := os.CreateTemp(t.TempDir(), "worker-*.js")
	assert.NoError(t, err)
	_, err = tmpFile.WriteString("addEventListener('fetch', e => {})")
	assert.NoError(t, err)
	tmpFile.Close()

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "deploy",
		"--name", "my-worker",
		"--file", tmpFile.Name(),
		"--json",
	)

	var worker WorkerSummary
	err = json.Unmarshal([]byte(output), &worker)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "my-worker", worker.ID)
}

func TestWorkersDeploy_FileMissing(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))

	err := runCmdErr(t, root, "workers", "deploy",
		"--name", "my-worker",
		"--file", "/nonexistent/path/worker.js",
	)
	assert.Error(t, err)
}

func TestWorkersDelete_Confirm_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "delete",
		"--name", "my-worker",
		"--confirm",
	)

	mustContain(t, output, "Deleted worker")
	mustContain(t, output, "my-worker")
}

func TestWorkersDelete_Confirm_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "delete",
		"--name", "my-worker",
		"--confirm",
		"--json",
	)

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "my-worker", result["name"])
}

func TestWorkersDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))
	output := runCmd(t, root, "workers", "delete",
		"--name", "my-worker",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-worker")
}

func TestWorkersDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkersTestCmd(factory))

	err := runCmdErr(t, root, "workers", "delete", "--name", "my-worker")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
