package fly

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newSecretsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("secrets",
		newSecretsListCmd(factory),
		newSecretsSetCmd(factory),
		newSecretsUnsetCmd(factory),
	)
}

func TestSecretsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "list", "--app", "my-app")

	mustContain(t, output, "API_KEY")
	mustContain(t, output, "abc123digest")
	mustContain(t, output, "DATABASE_URL")
	mustContain(t, output, "def456digest")
}

func TestSecretsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "list", "--app", "my-app", "--json")

	var results []SecretSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(results))
	}
	if results[0].Name != "API_KEY" {
		t.Errorf("expected first secret Name=API_KEY, got %s", results[0].Name)
	}
	if results[0].Digest != "abc123digest" {
		t.Errorf("expected first secret Digest=abc123digest, got %s", results[0].Digest)
	}
	if results[1].Name != "DATABASE_URL" {
		t.Errorf("expected second secret Name=DATABASE_URL, got %s", results[1].Name)
	}
}

func TestSecretsSet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "set", "--app", "my-app", "--key", "API_KEY", "--value", "super-secret")

	mustContain(t, output, "Set secret")
	mustContain(t, output, "API_KEY")
	mustContain(t, output, "my-app")
}

func TestSecretsSet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "set", "--app", "my-app", "--key", "API_KEY", "--value", "super-secret", "--json")

	var results []SecretSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 secrets after set, got %d", len(results))
	}
	if results[0].Name != "API_KEY" {
		t.Errorf("expected first secret Name=API_KEY, got %s", results[0].Name)
	}
}

func TestSecretsSet_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "set", "--app", "my-app", "--key", "API_KEY", "--value", "super-secret", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "API_KEY")
	mustContain(t, output, "my-app")
}

func TestSecretsSet_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "set", "--app", "my-app", "--key", "API_KEY", "--value", "super-secret", "--dry-run", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON dry-run output, got: %s\nerror: %v", output, err)
	}
	if result["action"] != "set" {
		t.Errorf("expected action=set, got %v", result["action"])
	}
	if result["key"] != "API_KEY" {
		t.Errorf("expected key=API_KEY, got %v", result["key"])
	}
}

func TestSecretsUnset_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "unset", "--app", "my-app", "--keys", "API_KEY")

	mustContain(t, output, "Unset secrets")
	mustContain(t, output, "my-app")
}

func TestSecretsUnset_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "unset", "--app", "my-app", "--keys", "API_KEY", "--json")

	var results []SecretSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	// Mock returns 1 secret remaining after unset
	if len(results) != 1 {
		t.Fatalf("expected 1 secret remaining after unset, got %d", len(results))
	}
	if results[0].Name != "DATABASE_URL" {
		t.Errorf("expected remaining secret Name=DATABASE_URL, got %s", results[0].Name)
	}
}

func TestSecretsUnset_MultipleKeys(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "unset", "--app", "my-app", "--keys", "API_KEY,DATABASE_URL")

	mustContain(t, output, "Unset secrets")
	mustContain(t, output, "my-app")
}

func TestSecretsUnset_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "unset", "--app", "my-app", "--keys", "API_KEY", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "API_KEY")
	mustContain(t, output, "my-app")
}

func TestSecretsUnset_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newSecretsTestCmd(factory))
	output := runCmd(t, root, "secrets", "unset", "--app", "my-app", "--keys", "API_KEY", "--dry-run", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON dry-run output, got: %s\nerror: %v", output, err)
	}
	if result["action"] != "unset" {
		t.Errorf("expected action=unset, got %v", result["action"])
	}
}
