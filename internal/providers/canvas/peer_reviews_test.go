package canvas

import (
	"strings"
	"testing"
)

func TestPeerReviewsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPeerReviewsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"peer-reviews", "list", "--course-id", "101", "--assignment-id", "501"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "43") {
		t.Errorf("expected assessor_id 43 in output, got: %s", output)
	}
	if !strings.Contains(output, "assigned") {
		t.Errorf("expected workflow_state 'assigned' in output, got: %s", output)
	}
}

func TestPeerReviewsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPeerReviewsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"peer-reviews", "list", "--course-id", "101", "--assignment-id", "501", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "assessor_id") {
		t.Errorf("JSON output should contain assessor_id field, got: %s", output)
	}
	if !strings.Contains(output, "workflow_state") {
		t.Errorf("JSON output should contain workflow_state field, got: %s", output)
	}
}

func TestPeerReviewsListMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPeerReviewsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"peer-reviews", "list", "--assignment-id", "501"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestPeerReviewsListMissingAssignmentID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPeerReviewsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"peer-reviews", "list", "--course-id", "101"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --assignment-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--assignment-id") {
		t.Errorf("error should mention --assignment-id, got: %v", execErr)
	}
}

func TestPeerReviewsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPeerReviewsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"peer-reviews", "create",
			"--course-id", "101",
			"--assignment-id", "501",
			"--user-id", "1",
			"--reviewer-id", "43",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "43") && !strings.Contains(output, "assigned") && !strings.Contains(output, "reviewer") {
		t.Errorf("expected peer review creation output, got: %s", output)
	}
}

func TestPeerReviewsDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPeerReviewsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"peer-reviews", "delete",
			"--course-id", "101",
			"--assignment-id", "501",
			"--user-id", "1",
			"--reviewer-id", "43",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "43") && !strings.Contains(output, "removed") {
		t.Errorf("expected peer review deletion output, got: %s", output)
	}
}
