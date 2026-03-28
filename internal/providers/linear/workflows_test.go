package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newWorkflowsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("workflows",
		newWorkflowsListCmd(factory),
	)
}

func TestWorkflowsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkflowsTestCmd(factory))
	output := runCmd(t, root, "workflows", "list", "--team", "team-abc1")

	mustContain(t, output, "state-abc1")
	mustContain(t, output, "Todo")
	mustContain(t, output, "unstarted")
	mustContain(t, output, "#e2e2e2")
	mustContain(t, output, "state-def2")
	mustContain(t, output, "In Progress")
	mustContain(t, output, "started")
	mustContain(t, output, "#f2c94c")
}

func TestWorkflowsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newWorkflowsTestCmd(factory))
	output := runCmd(t, root, "workflows", "list", "--team", "team-abc1", "--json")

	var results []WorkflowState
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "state-abc1", results[0].ID)
	assert.Equal(t, "Todo", results[0].Name)
	assert.Equal(t, "#e2e2e2", results[0].Color)
	assert.Equal(t, "unstarted", results[0].Type)
	assert.InDelta(t, 0.0, results[0].Position, 0.001)
	assert.Equal(t, "state-def2", results[1].ID)
	assert.Equal(t, "In Progress", results[1].Name)
	assert.Equal(t, "started", results[1].Type)
	assert.InDelta(t, 1.0, results[1].Position, 0.001)
}
