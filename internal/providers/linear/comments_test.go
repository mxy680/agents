package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newCommentsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("comments",
		newCommentsListCmd(factory),
		newCommentsCreateCmd(factory),
		newCommentsDeleteCmd(factory),
	)
}

func TestCommentsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "list", "--issue", "issue-abc1")

	mustContain(t, output, "cmt-abc1")
	mustContain(t, output, "First comment")
	mustContain(t, output, "Alice")
	mustContain(t, output, "cmt-def2")
	mustContain(t, output, "Second comment")
	mustContain(t, output, "Bob")
}

func TestCommentsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "list", "--issue", "issue-abc1", "--json")

	var results []CommentSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "cmt-abc1", results[0].ID)
	assert.Equal(t, "First comment", results[0].Body)
	assert.Equal(t, "Alice", results[0].User)
	assert.Equal(t, "2024-01-01T00:00:00Z", results[0].CreatedAt)
	assert.Equal(t, "cmt-def2", results[1].ID)
	assert.Equal(t, "Second comment", results[1].Body)
}

func TestCommentsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "create", "--issue", "issue-abc1", "--body", "Great work!")

	mustContain(t, output, "Created comment")
	mustContain(t, output, "cmt-new1")
}

func TestCommentsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "create", "--issue", "issue-abc1", "--body", "Great work!", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "cmt-new1", result["id"])
	assert.Equal(t, "Great work!", result["body"])
}

func TestCommentsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "create", "--issue", "issue-abc1", "--body", "Great work!", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "issue-abc1")
}

func TestCommentsDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "delete", "--id", "cmt-abc1", "--confirm")

	mustContain(t, output, "Deleted comment")
	mustContain(t, output, "cmt-abc1")
}

func TestCommentsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	err := runCmdErr(t, root, "comments", "delete", "--id", "cmt-abc1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestCommentsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "delete", "--id", "cmt-abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestCommentsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCommentsTestCmd(factory))
	output := runCmd(t, root, "comments", "delete", "--id", "cmt-abc1", "--confirm", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "cmt-abc1", result["id"])
}
