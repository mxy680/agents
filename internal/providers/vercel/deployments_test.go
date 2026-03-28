package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newDeploymentsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("deployments",
		newDeploymentsListCmd(factory),
		newDeploymentsGetCmd(factory),
		newDeploymentsCreateCmd(factory),
		newDeploymentsCancelCmd(factory),
		newDeploymentsDeleteCmd(factory),
	)
}

func TestDeploymentsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "list")

	mustContain(t, output, "dpl_abc123")
	mustContain(t, output, "READY")
	mustContain(t, output, "production")
	mustContain(t, output, "dpl_def456")
	mustContain(t, output, "BUILDING")
}

func TestDeploymentsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "list", "--json")

	var results []DeploymentSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 deployments, got %d", len(results))
	}
	if results[0].ID != "dpl_abc123" {
		t.Errorf("expected first deployment ID=dpl_abc123, got %s", results[0].ID)
	}
	if results[0].State != "READY" {
		t.Errorf("expected first deployment State=READY, got %s", results[0].State)
	}
	if results[0].Target != "production" {
		t.Errorf("expected first deployment Target=production, got %s", results[0].Target)
	}
	if results[1].ID != "dpl_def456" {
		t.Errorf("expected second deployment ID=dpl_def456, got %s", results[1].ID)
	}
	if results[1].State != "BUILDING" {
		t.Errorf("expected second deployment State=BUILDING, got %s", results[1].State)
	}
}

func TestDeploymentsList_WithFilters(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	// filters are passed as query params; mock returns all regardless
	output := runCmd(t, root, "deployments", "list", "--project", "prj_abc123", "--target", "production", "--state", "READY", "--json")

	var results []DeploymentSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 deployments, got %d", len(results))
	}
}

func TestDeploymentsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "get", "--id", "dpl_abc123", "--json")

	var detail DeploymentDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "dpl_abc123" {
		t.Errorf("expected ID=dpl_abc123, got %s", detail.ID)
	}
	if detail.State != "READY" {
		t.Errorf("expected State=READY, got %s", detail.State)
	}
	if detail.Creator != "alice" {
		t.Errorf("expected Creator=alice, got %s", detail.Creator)
	}
	if detail.GitBranch != "main" {
		t.Errorf("expected GitBranch=main, got %s", detail.GitBranch)
	}
	if detail.GitCommit != "abc123def456" {
		t.Errorf("expected GitCommit=abc123def456, got %s", detail.GitCommit)
	}
}

func TestDeploymentsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "get", "--id", "dpl_abc123")

	mustContain(t, output, "dpl_abc123")
	mustContain(t, output, "READY")
	mustContain(t, output, "alice")
	mustContain(t, output, "main")
}

func TestDeploymentsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "create", "--project", "my-nextjs-app", "--target", "production", "--json")

	var detail DeploymentDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "dpl_new789" {
		t.Errorf("expected ID=dpl_new789, got %s", detail.ID)
	}
}

func TestDeploymentsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "create", "--project", "my-nextjs-app", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-nextjs-app")
}

func TestDeploymentsCancel_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "cancel", "--id", "dpl_abc123", "--json")

	mustContain(t, output, "CANCELED")
}

func TestDeploymentsCancel_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "cancel", "--id", "dpl_abc123", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "dpl_abc123")
}

func TestDeploymentsDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	output := runCmd(t, root, "deployments", "delete", "--id", "dpl_abc123", "--confirm")

	mustContain(t, output, "Deleted")
	mustContain(t, output, "dpl_abc123")
}

func TestDeploymentsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDeploymentsTestCmd(factory))
	err := runCmdErr(t, root, "deployments", "delete", "--id", "dpl_abc123")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}
