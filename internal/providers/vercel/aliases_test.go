package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newAliasesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("aliases",
		newAliasesListCmd(factory),
		newAliasesAssignCmd(factory),
	)
}

func TestAliasesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAliasesTestCmd(factory))
	output := runCmd(t, root, "aliases", "list", "--deployment-id", "dpl_abc123", "--json")

	var aliases []AliasSummary
	if err := json.Unmarshal([]byte(output), &aliases); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(aliases) != 2 {
		t.Fatalf("expected 2 aliases, got %d", len(aliases))
	}
	if aliases[0].UID != "ali_abc1" {
		t.Errorf("expected first alias UID=ali_abc1, got %s", aliases[0].UID)
	}
	if aliases[0].Alias != "my-app.vercel.app" {
		t.Errorf("expected first alias Alias=my-app.vercel.app, got %s", aliases[0].Alias)
	}
	if aliases[0].DeploymentID != "dpl_abc123" {
		t.Errorf("expected first alias DeploymentID=dpl_abc123, got %s", aliases[0].DeploymentID)
	}
	if aliases[1].Alias != "my-app-custom.example.com" {
		t.Errorf("expected second alias Alias=my-app-custom.example.com, got %s", aliases[1].Alias)
	}
}

func TestAliasesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAliasesTestCmd(factory))
	output := runCmd(t, root, "aliases", "list", "--deployment-id", "dpl_abc123")

	mustContain(t, output, "ali_abc1")
	mustContain(t, output, "my-app.vercel.app")
}

func TestAliasesAssign_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAliasesTestCmd(factory))
	output := runCmd(t, root, "aliases", "assign",
		"--deployment-id", "dpl_abc123",
		"--alias", "my-new-alias.vercel.app",
		"--json",
	)

	var a AliasSummary
	if err := json.Unmarshal([]byte(output), &a); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if a.UID != "ali_new1" {
		t.Errorf("expected UID=ali_new1, got %s", a.UID)
	}
	if a.Alias != "my-new-alias.vercel.app" {
		t.Errorf("expected Alias=my-new-alias.vercel.app, got %s", a.Alias)
	}
}

func TestAliasesAssign_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAliasesTestCmd(factory))
	output := runCmd(t, root, "aliases", "assign",
		"--deployment-id", "dpl_abc123",
		"--alias", "my-new-alias.vercel.app",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-new-alias.vercel.app")
}
