package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newEnvTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("env",
		newEnvListCmd(factory),
		newEnvGetCmd(factory),
		newEnvSetCmd(factory),
		newEnvRemoveCmd(factory),
	)
}

func TestEnvList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))
	output := runCmd(t, root, "env", "list", "--project", "prj_abc123", "--json")

	var results []EnvSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(results))
	}
	if results[0].ID != "env_abc1" {
		t.Errorf("expected first env ID=env_abc1, got %s", results[0].ID)
	}
	if results[0].Key != "API_KEY" {
		t.Errorf("expected first env Key=API_KEY, got %s", results[0].Key)
	}
	if results[0].Type != "encrypted" {
		t.Errorf("expected first env Type=encrypted, got %s", results[0].Type)
	}
	if results[1].Key != "DATABASE_URL" {
		t.Errorf("expected second env Key=DATABASE_URL, got %s", results[1].Key)
	}
}

func TestEnvList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))
	output := runCmd(t, root, "env", "list", "--project", "prj_abc123")

	mustContain(t, output, "API_KEY")
	mustContain(t, output, "DATABASE_URL")
}

func TestEnvGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))
	output := runCmd(t, root, "env", "get", "--project", "prj_abc123", "--key", "env_abc1", "--json")

	var e EnvSummary
	if err := json.Unmarshal([]byte(output), &e); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if e.ID != "env_abc1" {
		t.Errorf("expected ID=env_abc1, got %s", e.ID)
	}
	if e.Key != "API_KEY" {
		t.Errorf("expected Key=API_KEY, got %s", e.Key)
	}
	if e.Value != "secret-value" {
		t.Errorf("expected Value=secret-value, got %s", e.Value)
	}
}

func TestEnvSet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))
	output := runCmd(t, root, "env", "set",
		"--project", "prj_abc123",
		"--key", "NEW_VAR",
		"--value", "my-value",
		"--json",
	)

	var e EnvSummary
	if err := json.Unmarshal([]byte(output), &e); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if e.ID != "env_created1" {
		t.Errorf("expected ID=env_created1, got %s", e.ID)
	}
	if e.Key != "NEW_VAR" {
		t.Errorf("expected Key=NEW_VAR, got %s", e.Key)
	}
}

func TestEnvSet_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))
	output := runCmd(t, root, "env", "set",
		"--project", "prj_abc123",
		"--key", "NEW_VAR",
		"--value", "my-value",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "NEW_VAR")
}

func TestEnvRemove_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))

	var execErr error
	output := captureStdout(t, func() {
		root.SetArgs([]string{"env", "remove", "--project", "prj_abc123", "--key", "env_abc1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	mustContain(t, output, "Deleted")
	mustContain(t, output, "env_abc1")
}

func TestEnvGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))
	output := runCmd(t, root, "env", "get", "--project", "prj_abc123", "--key", "env_abc1")

	mustContain(t, output, "API_KEY")
	mustContain(t, output, "secret-value")
}

func TestEnvRemove_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newEnvTestCmd(factory))
	err := runCmdErr(t, root, "env", "remove", "--project", "prj_abc123", "--key", "env_abc1")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}
