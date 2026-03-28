package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newLogsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("logs",
		newLogsGetCmd(factory),
	)
}

func TestLogsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLogsTestCmd(factory))
	output := runCmd(t, root, "logs", "get", "--deployment-id", "dpl_abc123", "--json")

	var events []LogEvent
	if err := json.Unmarshal([]byte(output), &events); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 log events, got %d", len(events))
	}
	if events[0].ID != "evt_abc1" {
		t.Errorf("expected first event ID=evt_abc1, got %s", events[0].ID)
	}
	if events[0].Text != "Build started" {
		t.Errorf("expected first event Text=Build started, got %s", events[0].Text)
	}
	if events[0].Type != "stdout" {
		t.Errorf("expected first event Type=stdout, got %s", events[0].Type)
	}
	if events[0].Source != "build" {
		t.Errorf("expected first event Source=build, got %s", events[0].Source)
	}
	if events[0].DeploymentID != "dpl_abc123" {
		t.Errorf("expected first event DeploymentID=dpl_abc123, got %s", events[0].DeploymentID)
	}
	if events[1].Text != "Build completed successfully" {
		t.Errorf("expected second event Text=Build completed successfully, got %s", events[1].Text)
	}
}

func TestLogsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newLogsTestCmd(factory))
	output := runCmd(t, root, "logs", "get", "--deployment-id", "dpl_abc123")

	mustContain(t, output, "Build started")
	mustContain(t, output, "Build completed successfully")
}
