package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newLabelsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("labels",
		newLabelsListCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsDeleteCmd(factory),
	)
}

func TestLabelsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "list")

	mustContain(t, output, "lbl-abc1")
	mustContain(t, output, "bug")
	mustContain(t, output, "#d73a4a")
	mustContain(t, output, "lbl-def2")
	mustContain(t, output, "enhancement")
	mustContain(t, output, "#a2eeef")
}

func TestLabelsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "list", "--json")

	var results []LabelSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "lbl-abc1", results[0].ID)
	assert.Equal(t, "bug", results[0].Name)
	assert.Equal(t, "#d73a4a", results[0].Color)
	assert.Equal(t, "lbl-def2", results[1].ID)
	assert.Equal(t, "enhancement", results[1].Name)
	assert.Equal(t, "#a2eeef", results[1].Color)
}

func TestLabelsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "create", "--name", "critical", "--color", "#ff0000")

	mustContain(t, output, "Created label")
	mustContain(t, output, "critical")
	mustContain(t, output, "#ff0000")
	mustContain(t, output, "lbl-new1")
}

func TestLabelsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "create", "--name", "critical", "--color", "#ff0000", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "lbl-new1", result["id"])
	assert.Equal(t, "critical", result["name"])
	assert.Equal(t, "#ff0000", result["color"])
}

func TestLabelsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "create", "--name", "critical", "--color", "#ff0000", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "critical")
}

func TestLabelsDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "delete", "--id", "lbl-abc1", "--confirm")

	mustContain(t, output, "Deleted label")
	mustContain(t, output, "lbl-abc1")
}

func TestLabelsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	err := runCmdErr(t, root, "labels", "delete", "--id", "lbl-abc1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestLabelsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "delete", "--id", "lbl-abc1", "--dry-run")

	mustContain(t, output, "DRY RUN")
}

func TestLabelsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLabelsTestCmd(factory))
	output := runCmd(t, root, "labels", "delete", "--id", "lbl-abc1", "--confirm", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "lbl-abc1", result["id"])
}
