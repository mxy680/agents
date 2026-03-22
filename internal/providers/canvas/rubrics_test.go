package canvas

import (
	"strings"
	"testing"
)

func TestRubricsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Essay Rubric") {
		t.Errorf("expected 'Essay Rubric' in output, got: %s", output)
	}
	if !strings.Contains(output, "Lab Rubric") {
		t.Errorf("expected 'Lab Rubric' in output, got: %s", output)
	}
	if !strings.Contains(output, "100") {
		t.Errorf("expected points in output, got: %s", output)
	}
}

func TestRubricsListMissingCourseID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestRubricsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "get", "--course-id", "101", "--rubric-id", "4001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Essay Rubric") {
		t.Errorf("expected rubric title, got: %s", output)
	}
	if !strings.Contains(output, "100") {
		t.Errorf("expected points in output, got: %s", output)
	}
}

func TestRubricsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "delete", "--course-id", "101", "--rubric-id", "4001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}
func TestRubricsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "create", "--course-id", "101", "--title", "New Rubric"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Rubric") && !strings.Contains(output, "4001") {
		t.Errorf("expected rubric creation output, got: %s", output)
	}
}

func TestRubricsUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "update", "--course-id", "101", "--rubric-id", "4001", "--title", "Updated Rubric"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "4001") || !strings.Contains(output, "updated") {
		t.Errorf("expected rubric update output, got: %s", output)
	}
}

func TestRubricsDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "delete", "--course-id", "101", "--rubric-id", "4001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "4001") || !strings.Contains(output, "deleted") {
		t.Errorf("expected rubric deletion output, got: %s", output)
	}
}

func TestRubricsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Essay Rubric") {
		t.Errorf("expected rubric name in JSON output, got: %s", output)
	}
}

func TestRubricsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "get", "--course-id", "101", "--rubric-id", "4001", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Essay Rubric") {
		t.Errorf("expected rubric name in JSON output, got: %s", output)
	}
}

func TestRubricsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"rubrics", "update", "--course-id", "101", "--rubric-id", "4001", "--title", "Updated Rubric", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "4001") {
		t.Errorf("expected rubric ID in JSON output, got: %s", output)
	}
}

func TestRubricsCreateWithPoints(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"rubrics", "create",
			"--course-id", "101",
			"--title", "New Rubric",
			"--points", "50",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "4001") && !strings.Contains(output, "New Rubric") && !strings.Contains(output, "created") {
		t.Errorf("expected rubric creation output, got: %s", output)
	}
}

func TestRubricsCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newRubricsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"rubrics", "create",
			"--course-id", "101",
			"--title", "New Rubric",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "4001") {
		t.Errorf("expected rubric ID in JSON output, got: %s", output)
	}
}
