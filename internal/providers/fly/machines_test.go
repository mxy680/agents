package fly

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newMachinesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("machines",
		newMachinesListCmd(factory),
		newMachinesGetCmd(factory),
		newMachinesCreateCmd(factory),
		newMachinesUpdateCmd(factory),
		newMachinesDeleteCmd(factory),
		newMachinesStartCmd(factory),
		newMachinesStopCmd(factory),
		newMachinesWaitCmd(factory),
	)
}

func TestMachinesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "list", "--app", "my-app")

	mustContain(t, output, "mach_abc1")
	mustContain(t, output, "machine-one")
	mustContain(t, output, "started")
	mustContain(t, output, "mach_def2")
	mustContain(t, output, "machine-two")
	mustContain(t, output, "stopped")
}

func TestMachinesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "list", "--app", "my-app", "--json")

	var results []MachineSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 machines, got %d", len(results))
	}
	if results[0].ID != "mach_abc1" {
		t.Errorf("expected first machine ID=mach_abc1, got %s", results[0].ID)
	}
	if results[0].Name != "machine-one" {
		t.Errorf("expected first machine Name=machine-one, got %s", results[0].Name)
	}
	if results[0].State != "started" {
		t.Errorf("expected first machine State=started, got %s", results[0].State)
	}
	if results[0].Image != "registry.fly.io/my-app:latest" {
		t.Errorf("expected first machine Image=registry.fly.io/my-app:latest, got %s", results[0].Image)
	}
	if results[1].Region != "lhr" {
		t.Errorf("expected second machine Region=lhr, got %s", results[1].Region)
	}
}

func TestMachinesGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "get", "--app", "my-app", "--machine", "mach_abc1")

	mustContain(t, output, "mach_abc1")
	mustContain(t, output, "machine-one")
	mustContain(t, output, "started")
	mustContain(t, output, "iad")
	mustContain(t, output, "inst_abc1")
	mustContain(t, output, "fdaa::1")
}

func TestMachinesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "get", "--app", "my-app", "--machine", "mach_abc1", "--json")

	var detail MachineDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "mach_abc1" {
		t.Errorf("expected ID=mach_abc1, got %s", detail.ID)
	}
	if detail.Name != "machine-one" {
		t.Errorf("expected Name=machine-one, got %s", detail.Name)
	}
	if detail.InstanceID != "inst_abc1" {
		t.Errorf("expected InstanceID=inst_abc1, got %s", detail.InstanceID)
	}
	if detail.PrivateIP != "fdaa::1" {
		t.Errorf("expected PrivateIP=fdaa::1, got %s", detail.PrivateIP)
	}
	if detail.Config == nil || detail.Config.Image != "registry.fly.io/my-app:latest" {
		t.Errorf("expected Config.Image=registry.fly.io/my-app:latest")
	}
}

func TestMachinesCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "create", "--app", "my-app", "--image", "registry.fly.io/my-app:v3")

	mustContain(t, output, "Created machine")
	mustContain(t, output, "mach_created1")
}

func TestMachinesCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "create", "--app", "my-app", "--image", "registry.fly.io/my-app:v3", "--json")

	var detail MachineDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "mach_created1" {
		t.Errorf("expected ID=mach_created1, got %s", detail.ID)
	}
	if detail.Config == nil || detail.Config.Image != "registry.fly.io/my-app:v3" {
		t.Errorf("expected Config.Image=registry.fly.io/my-app:v3")
	}
}

func TestMachinesCreate_WithRegionAndName(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "create",
		"--app", "my-app",
		"--image", "registry.fly.io/my-app:v3",
		"--region", "lhr",
		"--name", "custom-name",
	)

	mustContain(t, output, "Created machine")
}

func TestMachinesCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "create", "--app", "my-app", "--image", "registry.fly.io/my-app:v3", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-app")
}

func TestMachinesUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "update", "--app", "my-app", "--machine", "mach_abc1", "--image", "registry.fly.io/my-app:v4")

	mustContain(t, output, "Updated machine")
	mustContain(t, output, "mach_abc1")
}

func TestMachinesUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "update",
		"--app", "my-app",
		"--machine", "mach_abc1",
		"--image", "registry.fly.io/my-app:v4",
		"--json",
	)

	var detail MachineDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "mach_abc1" {
		t.Errorf("expected ID=mach_abc1, got %s", detail.ID)
	}
	if detail.Config == nil || detail.Config.Image != "registry.fly.io/my-app:v4" {
		t.Errorf("expected Config.Image=registry.fly.io/my-app:v4, got %+v", detail.Config)
	}
}

func TestMachinesUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "update",
		"--app", "my-app",
		"--machine", "mach_abc1",
		"--image", "registry.fly.io/my-app:v4",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "mach_abc1")
}

func TestMachinesDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "delete", "--app", "my-app", "--machine", "mach_abc1", "--confirm")

	mustContain(t, output, "Deleted machine")
	mustContain(t, output, "mach_abc1")
}

func TestMachinesDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	err := runCmdErr(t, root, "machines", "delete", "--app", "my-app", "--machine", "mach_abc1")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}

func TestMachinesDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "delete", "--app", "my-app", "--machine", "mach_abc1", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
	if result["machine"] != "mach_abc1" {
		t.Errorf("expected machine=mach_abc1, got %s", result["machine"])
	}
}

func TestMachinesDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "delete", "--app", "my-app", "--machine", "mach_abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "mach_abc1")
}

func TestMachinesStart_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "start", "--app", "my-app", "--machine", "mach_abc1")

	mustContain(t, output, "Started machine")
	mustContain(t, output, "mach_abc1")
}

func TestMachinesStart_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "start", "--app", "my-app", "--machine", "mach_abc1", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if result["status"] != "started" {
		t.Errorf("expected status=started, got %s", result["status"])
	}
	if result["machine"] != "mach_abc1" {
		t.Errorf("expected machine=mach_abc1, got %s", result["machine"])
	}
}

func TestMachinesStop_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "stop", "--app", "my-app", "--machine", "mach_abc1")

	mustContain(t, output, "Stopped machine")
	mustContain(t, output, "mach_abc1")
}

func TestMachinesStop_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "stop", "--app", "my-app", "--machine", "mach_abc1", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if result["status"] != "stopped" {
		t.Errorf("expected status=stopped, got %s", result["status"])
	}
}

func TestMachinesWait_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "wait", "--app", "my-app", "--machine", "mach_abc1", "--state", "started")

	mustContain(t, output, "mach_abc1")
	mustContain(t, output, "started")
}

func TestMachinesWait_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	output := runCmd(t, root, "machines", "wait", "--app", "my-app", "--machine", "mach_abc1", "--state", "started", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if result["machine"] != "mach_abc1" {
		t.Errorf("expected machine=mach_abc1, got %s", result["machine"])
	}
	if result["state"] != "started" {
		t.Errorf("expected state=started, got %s", result["state"])
	}
}

func TestMachinesWait_DefaultState(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newMachinesTestCmd(factory))
	// No --state flag — should default to "started"
	output := runCmd(t, root, "machines", "wait", "--app", "my-app", "--machine", "mach_abc1")

	mustContain(t, output, "mach_abc1")
	mustContain(t, output, "started")
}
