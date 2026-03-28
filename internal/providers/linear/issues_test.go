package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newIssuesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("issues",
		newIssuesListCmd(factory),
		newIssuesGetCmd(factory),
		newIssuesCreateCmd(factory),
		newIssuesUpdateCmd(factory),
		newIssuesDeleteCmd(factory),
	)
}

func TestIssuesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "list", "--team", "team-abc1")

	mustContain(t, output, "ENG-1")
	mustContain(t, output, "Fix login bug")
	mustContain(t, output, "ENG-2")
	mustContain(t, output, "Add dark mode")
	mustContain(t, output, "In Progress")
}

func TestIssuesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "list", "--team", "team-abc1", "--json")

	var results []IssueSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "issue-abc1", results[0].ID)
	assert.Equal(t, "ENG-1", results[0].Identifier)
	assert.Equal(t, "Fix login bug", results[0].Title)
	assert.Equal(t, "In Progress", results[0].State)
	assert.Equal(t, 2, results[0].Priority)
	assert.Equal(t, "Alice", results[0].Assignee)
	assert.Equal(t, "issue-def2", results[1].ID)
	assert.Equal(t, "ENG-2", results[1].Identifier)
	assert.Equal(t, "", results[1].Assignee)
}

func TestIssuesList_WithLimit(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "list", "--team", "team-abc1", "--limit", "10")

	mustContain(t, output, "ENG-1")
}

func TestIssuesGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "get", "--id", "issue-abc1")

	mustContain(t, output, "issue-abc1")
	mustContain(t, output, "ENG-1")
	mustContain(t, output, "Fix login bug")
	mustContain(t, output, "In Progress")
	mustContain(t, output, "Alice")
	mustContain(t, output, "Engineering")
	mustContain(t, output, "bug")
	mustContain(t, output, "Users can't log in with SSO")
}

func TestIssuesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "get", "--id", "issue-abc1", "--json")

	var detail IssueDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "issue-abc1", detail.ID)
	assert.Equal(t, "ENG-1", detail.Identifier)
	assert.Equal(t, "Fix login bug", detail.Title)
	assert.Equal(t, "Users can't log in with SSO", detail.Description)
	assert.Equal(t, "In Progress", detail.State)
	assert.Equal(t, 2, detail.Priority)
	assert.Equal(t, "Alice", detail.Assignee)
	assert.Equal(t, "Engineering", detail.Team)
	assert.Contains(t, detail.Labels, "bug")
}

func TestIssuesCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "create", "--team", "team-abc1", "--title", "New bug report")

	mustContain(t, output, "Created issue")
	mustContain(t, output, "New bug report")
}

func TestIssuesCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "create", "--team", "team-abc1", "--title", "New bug report", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "issue-new1", result["id"])
	assert.Equal(t, "New bug report", result["title"])
}

func TestIssuesCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "create", "--team", "team-abc1", "--title", "New bug report", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "New bug report")
}

func TestIssuesCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "create", "--team", "team-abc1", "--title", "New bug report", "--dry-run", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON dry-run output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "create", result["action"])
}

func TestIssuesUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "update", "--id", "issue-abc1", "--title", "Updated issue title")

	mustContain(t, output, "Updated issue")
	mustContain(t, output, "ENG-1")
}

func TestIssuesUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "update", "--id", "issue-abc1", "--title", "Updated issue title", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "issue-abc1", result["id"])
	assert.Equal(t, "Updated issue title", result["title"])
}

func TestIssuesUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "update", "--id", "issue-abc1", "--title", "New title", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestIssuesDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "delete", "--id", "issue-abc1", "--confirm")

	mustContain(t, output, "Deleted issue")
	mustContain(t, output, "issue-abc1")
}

func TestIssuesDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	err := runCmdErr(t, root, "issues", "delete", "--id", "issue-abc1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestIssuesDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "delete", "--id", "issue-abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestIssuesDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))
	output := runCmd(t, root, "issues", "delete", "--id", "issue-abc1", "--confirm", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "issue-abc1", result["id"])
}
