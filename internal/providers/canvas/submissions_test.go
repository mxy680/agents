package canvas

import (
	"strings"
	"testing"
)

func TestSubmissionsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSubmissionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"submissions", "list", "--course-id", "101", "--assignment-id", "501"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "graded") {
		t.Errorf("output should contain workflow state, got: %s", output)
	}
	if !strings.Contains(output, "A") {
		t.Errorf("output should contain grade, got: %s", output)
	}
	if !strings.Contains(output, "2026-02-01") {
		t.Errorf("output should contain submitted_at date, got: %s", output)
	}
}

func TestSubmissionsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSubmissionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"submissions", "list", "--course-id", "101", "--assignment-id", "501", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"workflow_state"`) {
		t.Errorf("JSON output should contain workflow_state field, got: %s", output)
	}
	if !strings.Contains(output, `"grade"`) {
		t.Errorf("JSON output should contain grade field, got: %s", output)
	}
	if !strings.Contains(output, "graded") {
		t.Errorf("JSON output should contain workflow state value, got: %s", output)
	}
}

func TestSubmissionsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSubmissionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"submissions", "get", "--course-id", "101", "--assignment-id", "501", "--user-id", "1"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "graded") {
		t.Errorf("output should contain workflow state, got: %s", output)
	}
	if !strings.Contains(output, "A") {
		t.Errorf("output should contain grade, got: %s", output)
	}
	if !strings.Contains(output, "95") {
		t.Errorf("output should contain score, got: %s", output)
	}
	if !strings.Contains(output, "2026-02-01") {
		t.Errorf("output should contain submitted_at date, got: %s", output)
	}
}

func TestSubmissionsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSubmissionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"submissions", "get", "--course-id", "101", "--assignment-id", "501", "--user-id", "1", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"workflow_state"`) {
		t.Errorf("JSON output should contain workflow_state field, got: %s", output)
	}
	if !strings.Contains(output, `"grade"`) {
		t.Errorf("JSON output should contain grade field, got: %s", output)
	}
	if !strings.Contains(output, `"score"`) {
		t.Errorf("JSON output should contain score field, got: %s", output)
	}
}

func TestSubmissionsGetMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSubmissionsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"submissions", "get", "--assignment-id", "501"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}
