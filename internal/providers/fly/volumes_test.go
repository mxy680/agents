package fly

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newVolumesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("volumes",
		newVolumesListCmd(factory),
		newVolumesGetCmd(factory),
		newVolumesCreateCmd(factory),
		newVolumesExtendCmd(factory),
		newVolumesDeleteCmd(factory),
		newVolumesSnapshotsCmd(factory),
	)
}

func TestVolumesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "list", "--app", "my-app")

	mustContain(t, output, "vol_abc1")
	mustContain(t, output, "my-volume")
	mustContain(t, output, "iad")
	mustContain(t, output, "vol_def2")
	mustContain(t, output, "my-other-volume")
	mustContain(t, output, "lhr")
}

func TestVolumesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "list", "--app", "my-app", "--json")

	var results []VolumeSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(results))
	}
	if results[0].ID != "vol_abc1" {
		t.Errorf("expected first volume ID=vol_abc1, got %s", results[0].ID)
	}
	if results[0].Name != "my-volume" {
		t.Errorf("expected first volume Name=my-volume, got %s", results[0].Name)
	}
	if results[0].SizeGB != 10 {
		t.Errorf("expected first volume SizeGB=10, got %d", results[0].SizeGB)
	}
	if results[1].SizeGB != 20 {
		t.Errorf("expected second volume SizeGB=20, got %d", results[1].SizeGB)
	}
}

func TestVolumesGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "get", "--app", "my-app", "--volume", "vol_abc1")

	mustContain(t, output, "vol_abc1")
	mustContain(t, output, "my-volume")
	mustContain(t, output, "iad")
	mustContain(t, output, "10")
	mustContain(t, output, "true") // encrypted
	mustContain(t, output, "mach_abc1")
}

func TestVolumesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "get", "--app", "my-app", "--volume", "vol_abc1", "--json")

	var detail VolumeDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "vol_abc1" {
		t.Errorf("expected ID=vol_abc1, got %s", detail.ID)
	}
	if detail.Name != "my-volume" {
		t.Errorf("expected Name=my-volume, got %s", detail.Name)
	}
	if !detail.Encrypted {
		t.Error("expected Encrypted=true")
	}
	if detail.AttachedMachineID != "mach_abc1" {
		t.Errorf("expected AttachedMachineID=mach_abc1, got %s", detail.AttachedMachineID)
	}
	if detail.SizeGB != 10 {
		t.Errorf("expected SizeGB=10, got %d", detail.SizeGB)
	}
}

func TestVolumesCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "create",
		"--app", "my-app",
		"--name", "new-volume",
		"--region", "iad",
		"--size", "5",
	)

	mustContain(t, output, "Created volume")
	mustContain(t, output, "new-volume")
}

func TestVolumesCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "create",
		"--app", "my-app",
		"--name", "new-volume",
		"--region", "iad",
		"--size", "5",
		"--json",
	)

	var detail VolumeDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "vol_created1" {
		t.Errorf("expected ID=vol_created1, got %s", detail.ID)
	}
	if detail.Name != "new-volume" {
		t.Errorf("expected Name=new-volume, got %s", detail.Name)
	}
	if detail.SizeGB != 5 {
		t.Errorf("expected SizeGB=5, got %d", detail.SizeGB)
	}
}

func TestVolumesCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "create",
		"--app", "my-app",
		"--name", "new-volume",
		"--region", "iad",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "new-volume")
	mustContain(t, output, "iad")
}

func TestVolumesCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "create",
		"--app", "my-app",
		"--name", "new-volume",
		"--region", "iad",
		"--dry-run",
		"--json",
	)

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON dry-run output, got: %s\nerror: %v", output, err)
	}
	if result["name"] != "new-volume" {
		t.Errorf("expected name=new-volume, got %v", result["name"])
	}
}

func TestVolumesExtend_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "extend", "--app", "my-app", "--volume", "vol_abc1", "--size", "20")

	mustContain(t, output, "Extended volume")
	mustContain(t, output, "vol_abc1")
	mustContain(t, output, "20")
}

func TestVolumesExtend_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "extend", "--app", "my-app", "--volume", "vol_abc1", "--size", "20", "--json")

	var detail VolumeDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.ID != "vol_abc1" {
		t.Errorf("expected ID=vol_abc1, got %s", detail.ID)
	}
	if detail.SizeGB != 20 {
		t.Errorf("expected SizeGB=20, got %d", detail.SizeGB)
	}
}

func TestVolumesExtend_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "extend", "--app", "my-app", "--volume", "vol_abc1", "--size", "20", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "vol_abc1")
	mustContain(t, output, "20")
}

func TestVolumesDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "delete", "--app", "my-app", "--volume", "vol_abc1", "--confirm")

	mustContain(t, output, "Deleted volume")
	mustContain(t, output, "vol_abc1")
}

func TestVolumesDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	err := runCmdErr(t, root, "volumes", "delete", "--app", "my-app", "--volume", "vol_abc1")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}

func TestVolumesDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "delete", "--app", "my-app", "--volume", "vol_abc1", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
	if result["volume"] != "vol_abc1" {
		t.Errorf("expected volume=vol_abc1, got %s", result["volume"])
	}
}

func TestVolumesDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "delete", "--app", "my-app", "--volume", "vol_abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "vol_abc1")
}

func TestVolumesSnapshots_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "snapshots", "--app", "my-app", "--volume", "vol_abc1")

	mustContain(t, output, "snap_abc1")
	mustContain(t, output, "complete")
	mustContain(t, output, "snap_def2")
}

func TestVolumesSnapshots_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newVolumesTestCmd(factory))
	output := runCmd(t, root, "volumes", "snapshots", "--app", "my-app", "--volume", "vol_abc1", "--json")

	var snapshots []VolumeSnapshot
	if err := json.Unmarshal([]byte(output), &snapshots); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
	if snapshots[0].ID != "snap_abc1" {
		t.Errorf("expected first snapshot ID=snap_abc1, got %s", snapshots[0].ID)
	}
	if snapshots[0].Status != "complete" {
		t.Errorf("expected first snapshot Status=complete, got %s", snapshots[0].Status)
	}
	if snapshots[0].SizeBytes != 1073741824 {
		t.Errorf("expected first snapshot SizeBytes=1073741824, got %d", snapshots[0].SizeBytes)
	}
}
