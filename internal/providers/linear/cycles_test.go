package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newCyclesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("cycles",
		newCyclesListCmd(factory),
		newCyclesGetCmd(factory),
		newCyclesCurrentCmd(factory),
	)
}

func TestCyclesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCyclesTestCmd(factory))
	output := runCmd(t, root, "cycles", "list", "--team", "team-abc1")

	mustContain(t, output, "cycle-abc1")
	mustContain(t, output, "cycle-def2")
	mustContain(t, output, "2024-01-01")
	mustContain(t, output, "2024-01-14")
}

func TestCyclesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCyclesTestCmd(factory))
	output := runCmd(t, root, "cycles", "list", "--team", "team-abc1", "--json")

	var results []CycleSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "cycle-abc1", results[0].ID)
	assert.Equal(t, 1, results[0].Number)
	assert.Equal(t, "2024-01-01T00:00:00Z", results[0].StartsAt)
	assert.Equal(t, "2024-01-14T00:00:00Z", results[0].EndsAt)
	assert.Equal(t, "cycle-def2", results[1].ID)
	assert.Equal(t, 2, results[1].Number)
}

func TestCyclesGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCyclesTestCmd(factory))
	output := runCmd(t, root, "cycles", "get", "--id", "cycle-abc1")

	mustContain(t, output, "cycle-abc1")
	mustContain(t, output, "3")
	mustContain(t, output, "2024-01-01")
	mustContain(t, output, "2024-01-14")
	mustContain(t, output, "Issues:")
}

func TestCyclesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCyclesTestCmd(factory))
	output := runCmd(t, root, "cycles", "get", "--id", "cycle-abc1", "--json")

	var detail CycleDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "cycle-abc1", detail.ID)
	assert.Equal(t, 3, detail.Number)
	assert.Equal(t, "2024-01-01T00:00:00Z", detail.StartsAt)
	assert.Equal(t, "2024-01-14T00:00:00Z", detail.EndsAt)
	assert.Len(t, detail.Issues, 1)
	assert.Equal(t, "issue-abc1", detail.Issues[0].ID)
	assert.Equal(t, "Fix login bug", detail.Issues[0].Title)
}

func TestCyclesCurrent_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCyclesTestCmd(factory))
	output := runCmd(t, root, "cycles", "current", "--team", "team-abc1")

	mustContain(t, output, "cycle-abc1")
	mustContain(t, output, "3")
	mustContain(t, output, "2024-01-01")
}

func TestCyclesCurrent_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCyclesTestCmd(factory))
	output := runCmd(t, root, "cycles", "current", "--team", "team-abc1", "--json")

	var summary CycleSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "cycle-abc1", summary.ID)
	assert.Equal(t, 3, summary.Number)
}
