package fly

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newAppsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("apps",
		newAppsListCmd(factory),
		newAppsGetCmd(factory),
		newAppsCreateCmd(factory),
		newAppsDeleteCmd(factory),
	)
}

func TestAppsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "list")

	mustContain(t, output, "my-app")
	mustContain(t, output, "my-other-app")
	mustContain(t, output, "running")
	mustContain(t, output, "suspended")
}

func TestAppsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "list", "--json")

	var results []AppSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(results))
	}
	if results[0].ID != "app_abc1" {
		t.Errorf("expected first app ID=app_abc1, got %s", results[0].ID)
	}
	if results[0].Name != "my-app" {
		t.Errorf("expected first app Name=my-app, got %s", results[0].Name)
	}
	if results[0].Status != "running" {
		t.Errorf("expected first app Status=running, got %s", results[0].Status)
	}
	if results[1].Name != "my-other-app" {
		t.Errorf("expected second app Name=my-other-app, got %s", results[1].Name)
	}
}

func TestAppsList_WithOrgFilter(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "list", "--org", "my-org")

	// The mock returns the same response regardless of query param; just verify it runs
	mustContain(t, output, "my-app")
}

func TestAppsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "get", "--app", "my-app")

	mustContain(t, output, "app_abc1")
	mustContain(t, output, "my-app")
	mustContain(t, output, "running")
	mustContain(t, output, "my-org")
	mustContain(t, output, "my-app.fly.dev")
}

func TestAppsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "get", "--app", "my-app", "--json")

	var detail AppDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "app_abc1" {
		t.Errorf("expected ID=app_abc1, got %s", detail.ID)
	}
	if detail.Name != "my-app" {
		t.Errorf("expected Name=my-app, got %s", detail.Name)
	}
	if detail.Hostname != "my-app.fly.dev" {
		t.Errorf("expected Hostname=my-app.fly.dev, got %s", detail.Hostname)
	}
	if detail.AppURL != "https://my-app.fly.dev" {
		t.Errorf("expected AppURL=https://my-app.fly.dev, got %s", detail.AppURL)
	}
}

func TestAppsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "create", "--name", "new-app", "--org", "my-org")

	mustContain(t, output, "Created app")
	mustContain(t, output, "new-app")
}

func TestAppsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "create", "--name", "new-app", "--org", "my-org", "--json")

	var detail AppDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "app_created1" {
		t.Errorf("expected ID=app_created1, got %s", detail.ID)
	}
	if detail.Name != "new-app" {
		t.Errorf("expected Name=new-app, got %s", detail.Name)
	}
}

func TestAppsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "create", "--name", "new-app", "--org", "my-org", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "new-app")
	mustContain(t, output, "my-org")
}

func TestAppsCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "create", "--name", "new-app", "--org", "my-org", "--dry-run", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON dry-run output, got: %s\nerror: %v", output, err)
	}
	if result["action"] != "create" {
		t.Errorf("expected action=create, got %v", result["action"])
	}
}

func TestAppsDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "delete", "--app", "my-app", "--confirm")

	mustContain(t, output, "Deleted app")
	mustContain(t, output, "my-app")
}

func TestAppsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	err := runCmdErr(t, root, "apps", "delete", "--app", "my-app")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}

func TestAppsDelete_Force(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "delete", "--app", "my-app", "--confirm", "--force")

	mustContain(t, output, "Deleted app")
}

func TestAppsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "delete", "--app", "my-app", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
	if result["app"] != "my-app" {
		t.Errorf("expected app=my-app, got %s", result["app"])
	}
}

func TestAppsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAppsTestCmd(factory))
	output := runCmd(t, root, "apps", "delete", "--app", "my-app", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-app")
}
