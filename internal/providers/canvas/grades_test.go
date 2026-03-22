package canvas

import (
	"strings"
	"testing"
)

func TestGradesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGradesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"grades", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// The enrollment mock returns grade data with user IDs.
	if !strings.Contains(output, "42") {
		t.Errorf("expected user_id 42 in grades output, got: %s", output)
	}
}

func TestGradesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGradesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"grades", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"user_id"`) {
		t.Errorf("JSON output should contain user_id field, got: %s", output)
	}
	if !strings.Contains(output, `"enrollment_state"`) {
		t.Errorf("JSON output should contain enrollment_state field, got: %s", output)
	}
}

func TestGradesListMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGradesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"grades", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestGradesHistoryText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGradesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"grades", "history", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "grade_change") && !strings.Contains(output, "events") {
		t.Errorf("expected grade change events in output, got: %s", output)
	}
}

func TestGradesHistoryMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newGradesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"grades", "history"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}
