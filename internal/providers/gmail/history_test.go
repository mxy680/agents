package gmail

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHistoryListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestHistoryCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"history", "list", "--start-history-id=12300", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("history list --json failed: %v", execErr)
	}

	var result HistoryResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	if len(result.History) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(result.History))
	}

	// First entry: one message added
	first := result.History[0]
	if first.ID != 12345 {
		t.Errorf("expected first entry ID=12345, got %d", first.ID)
	}
	if len(first.MessagesAdded) != 1 || first.MessagesAdded[0] != "msg1" {
		t.Errorf("expected first entry to have msg1 added, got %v", first.MessagesAdded)
	}

	// Second entry: one label added
	second := result.History[1]
	if second.ID != 12346 {
		t.Errorf("expected second entry ID=12346, got %d", second.ID)
	}
	if len(second.LabelsAdded) != 1 {
		t.Errorf("expected second entry to have 1 label change, got %d", len(second.LabelsAdded))
	}

	if result.HistoryID != 12347 {
		t.Errorf("expected historyId=12347, got %d", result.HistoryID)
	}
}

func TestHistoryListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestHistoryCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"history", "list", "--start-history-id=12300"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("history list (text) failed: %v", execErr)
	}

	// Verify the header row is present and both history IDs appear.
	if !strings.Contains(output, "HISTORY_ID") {
		t.Errorf("expected HISTORY_ID header in text output, got:\n%s", output)
	}
	if !strings.Contains(output, "12345") {
		t.Errorf("expected history ID 12345 in text output, got:\n%s", output)
	}
	if !strings.Contains(output, "12346") {
		t.Errorf("expected history ID 12346 in text output, got:\n%s", output)
	}
}
