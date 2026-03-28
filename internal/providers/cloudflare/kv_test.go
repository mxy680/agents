package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newKVTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("kv",
		newKVNamespacesListCmd(factory),
		newKVNamespacesCreateCmd(factory),
		newKVKeysListCmd(factory),
		newKVGetCmd(factory),
		newKVPutCmd(factory),
		newKVDeleteCmd(factory),
	)
}

func TestKVNamespacesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "namespaces-list")

	mustContain(t, output, "kv_abc1")
	mustContain(t, output, "MY_KV_NS")
	mustContain(t, output, "kv_def2")
	mustContain(t, output, "ANOTHER_KV_NS")
}

func TestKVNamespacesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "namespaces-list", "--json")

	var namespaces []KVNamespaceSummary
	err := json.Unmarshal([]byte(output), &namespaces)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, namespaces, 2)
	assert.Equal(t, "kv_abc1", namespaces[0].ID)
	assert.Equal(t, "MY_KV_NS", namespaces[0].Title)
	assert.Equal(t, "kv_def2", namespaces[1].ID)
}

func TestKVNamespacesCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "namespaces-create", "--title", "NEW_KV_NS")

	mustContain(t, output, "Created KV namespace")
	mustContain(t, output, "NEW_KV_NS")
}

func TestKVNamespacesCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "namespaces-create", "--title", "NEW_KV_NS", "--json")

	var ns KVNamespaceSummary
	err := json.Unmarshal([]byte(output), &ns)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "kv_new1", ns.ID)
	assert.Equal(t, "NEW_KV_NS", ns.Title)
}

func TestKVNamespacesCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "namespaces-create", "--title", "NEW_KV_NS", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "NEW_KV_NS")
}

func TestKVKeysList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "keys-list", "--namespace", "kv_abc1")

	mustContain(t, output, "key-one")
	mustContain(t, output, "key-two")
}

func TestKVKeysList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "keys-list", "--namespace", "kv_abc1", "--json")

	var keys []KVKeySummary
	err := json.Unmarshal([]byte(output), &keys)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, keys, 2)
	assert.Equal(t, "key-one", keys[0].Name)
	assert.Equal(t, int64(0), keys[0].Expiration)
	assert.Equal(t, "key-two", keys[1].Name)
	assert.Equal(t, int64(9999999999), keys[1].Expiration)
}

func TestKVGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "get", "--namespace", "kv_abc1", "--key", "my-key")

	mustContain(t, output, "hello-world")
}

func TestKVGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "get", "--namespace", "kv_abc1", "--key", "my-key", "--json")

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "my-key", result["key"])
	assert.Equal(t, "hello-world", result["value"])
}

func TestKVPut_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "put",
		"--namespace", "kv_abc1",
		"--key", "my-key",
		"--value", "hello-world",
	)

	mustContain(t, output, "Stored key")
	mustContain(t, output, "my-key")
}

func TestKVPut_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "put",
		"--namespace", "kv_abc1",
		"--key", "my-key",
		"--value", "hello-world",
		"--json",
	)

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "ok", result["status"])
	assert.Equal(t, "my-key", result["key"])
}

func TestKVPut_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "put",
		"--namespace", "kv_abc1",
		"--key", "my-key",
		"--value", "hello-world",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-key")
}

func TestKVDelete_Confirm_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "delete",
		"--namespace", "kv_abc1",
		"--key", "my-key",
		"--confirm",
	)

	mustContain(t, output, "Deleted key")
	mustContain(t, output, "my-key")
}

func TestKVDelete_Confirm_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "delete",
		"--namespace", "kv_abc1",
		"--key", "my-key",
		"--confirm",
		"--json",
	)

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "my-key", result["key"])
}

func TestKVDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))
	output := runCmd(t, root, "kv", "delete",
		"--namespace", "kv_abc1",
		"--key", "my-key",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-key")
}

func TestKVDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newKVTestCmd(factory))

	err := runCmdErr(t, root, "kv", "delete",
		"--namespace", "kv_abc1",
		"--key", "my-key",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
