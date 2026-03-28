package vercel

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newWebhooksTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("webhooks",
		newWebhooksListCmd(factory),
		newWebhooksCreateCmd(factory),
		newWebhooksDeleteCmd(factory),
	)
}

func TestWebhooksList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "list", "--json")

	var hooks []WebhookSummary
	if err := json.Unmarshal([]byte(output), &hooks); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(hooks) != 2 {
		t.Fatalf("expected 2 webhooks, got %d", len(hooks))
	}
	if hooks[0].ID != "hook_abc1" {
		t.Errorf("expected first webhook ID=hook_abc1, got %s", hooks[0].ID)
	}
	if hooks[0].URL != "https://hooks.example.com/vercel" {
		t.Errorf("expected first webhook URL=https://hooks.example.com/vercel, got %s", hooks[0].URL)
	}
	if len(hooks[0].Events) < 2 {
		t.Errorf("expected at least 2 events on first webhook, got %d", len(hooks[0].Events))
	}
	found := false
	for _, e := range hooks[0].Events {
		if e == "deployment.created" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected deployment.created in first webhook events, got %v", hooks[0].Events)
	}
}

func TestWebhooksList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "list")

	mustContain(t, output, "hook_abc1")
	mustContain(t, output, "hooks.example.com")
	mustContain(t, output, "deployment.created")
}

func TestWebhooksCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "create",
		"--url", "https://hooks.example.com/new",
		"--events", "deployment.created,deployment.error",
		"--json",
	)

	var h WebhookSummary
	if err := json.Unmarshal([]byte(output), &h); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if h.ID != "hook_new1" {
		t.Errorf("expected ID=hook_new1, got %s", h.ID)
	}
	if h.URL != "https://hooks.example.com/new" {
		t.Errorf("expected URL=https://hooks.example.com/new, got %s", h.URL)
	}
}

func TestWebhooksCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	output := runCmd(t, root, "webhooks", "create",
		"--url", "https://hooks.example.com/new",
		"--events", "deployment.created",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "hooks.example.com")
}

func TestWebhooksDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))

	var execErr error
	output := captureStdout(t, func() {
		root.SetArgs([]string{"webhooks", "delete", "--id", "hook_abc1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	mustContain(t, output, "Deleted")
	mustContain(t, output, "hook_abc1")
}

func TestWebhooksDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWebhooksTestCmd(factory))
	err := runCmdErr(t, root, "webhooks", "delete", "--id", "hook_abc1")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	if !strings.Contains(err.Error(), "irreversible") {
		t.Errorf("expected error to contain 'irreversible', got: %s", err.Error())
	}
}
