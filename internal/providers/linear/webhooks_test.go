package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newWebhooksTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("webhooks",
		newWebhooksListCmd(factory),
		newWebhooksCreateCmd(factory),
		newWebhooksDeleteCmd(factory),
	)
}

func TestWebhooksList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "list")

	mustContain(t, output, "wh-abc1")
	mustContain(t, output, "https://hooks.example.com/linear")
	mustContain(t, output, "Engineering")
	mustContain(t, output, "wh-def2")
	mustContain(t, output, "https://hooks.example.com/linear2")
}

func TestWebhooksList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "list", "--json")

	var results []WebhookSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "wh-abc1", results[0].ID)
	assert.Equal(t, "https://hooks.example.com/linear", results[0].URL)
	assert.True(t, results[0].Enabled)
	assert.Equal(t, "Engineering", results[0].Team)
	assert.Equal(t, "2024-01-01T00:00:00Z", results[0].CreatedAt)
	assert.Equal(t, "wh-def2", results[1].ID)
	assert.False(t, results[1].Enabled)
	assert.Equal(t, "", results[1].Team)
}

func TestWebhooksCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "create",
		"--url", "https://hooks.example.com/new",
		"--team", "team-abc1",
	)

	mustContain(t, output, "Created webhook")
	mustContain(t, output, "https://hooks.example.com/new")
	mustContain(t, output, "wh-new1")
}

func TestWebhooksCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "create",
		"--url", "https://hooks.example.com/new",
		"--team", "team-abc1",
		"--json",
	)

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "wh-new1", result["id"])
	assert.Equal(t, "https://hooks.example.com/new", result["url"])
}

func TestWebhooksCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "create",
		"--url", "https://hooks.example.com/new",
		"--team", "team-abc1",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "https://hooks.example.com/new")
}

func TestWebhooksCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "create",
		"--url", "https://hooks.example.com/new",
		"--team", "team-abc1",
		"--dry-run", "--json",
	)

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON dry-run output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "create", result["action"])
}

func TestWebhooksDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "delete", "--id", "wh-abc1", "--confirm")

	mustContain(t, output, "Deleted webhook")
	mustContain(t, output, "wh-abc1")
}

func TestWebhooksDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	err := runCmdErr(t, root, "webhooks", "delete", "--id", "wh-abc1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestWebhooksDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "delete", "--id", "wh-abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestWebhooksDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "delete", "--id", "wh-abc1", "--confirm", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "wh-abc1", result["id"])
}
