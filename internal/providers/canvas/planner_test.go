package canvas

import (
	"strings"
	"testing"
)

func TestPlannerListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Homework 1") {
		t.Errorf("expected 'Homework 1' planner item, got: %s", output)
	}
	if !strings.Contains(output, "assignment") {
		t.Errorf("expected plannable_type 'assignment', got: %s", output)
	}
}

func TestPlannerNotesText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "notes"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Study for Exam") {
		t.Errorf("expected 'Study for Exam' note, got: %s", output)
	}
	if !strings.Contains(output, "Review Notes") {
		t.Errorf("expected 'Review Notes' note, got: %s", output)
	}
}

func TestPlannerCreateNoteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "create-note", "--title", "New Study Session", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
}

func TestPlannerDeleteNoteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"planner", "delete-note", "--note-id", "7001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestPlannerCreateNoteLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "create-note", "--title", "New Study Session"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "Study for Exam") && !strings.Contains(output, "7001") {
		t.Errorf("expected note creation output, got: %s", output)
	}
}

func TestPlannerUpdateNoteLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "update-note", "--note-id", "7001", "--title", "Updated Study Plan"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "7001") || !strings.Contains(output, "updated") {
		t.Errorf("expected note update output, got: %s", output)
	}
}

func TestPlannerDeleteNoteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "delete-note", "--note-id", "7001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "7001") || !strings.Contains(output, "deleted") {
		t.Errorf("expected note deletion output, got: %s", output)
	}
}

func TestPlannerOverridesText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "overrides"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "assignment") && !strings.Contains(output, "30001") && !strings.Contains(output, "override") {
		t.Errorf("expected overrides output, got: %s", output)
	}
}

func TestPlannerOverrideLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"planner", "override",
			"--plannable-type", "assignment",
			"--plannable-id", "501",
			"--marked-complete",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "override") && !strings.Contains(output, "501") && !strings.Contains(output, "30001") {
		t.Errorf("expected override output, got: %s", output)
	}
}

func TestPlannerListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "plannable_type") {
		t.Errorf("JSON output should contain plannable_type field, got: %s", output)
	}
}

func TestPlannerNotesJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "notes", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Study for Exam") {
		t.Errorf("expected note in JSON output, got: %s", output)
	}
}

func TestPlannerOverridesJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "overrides", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "plannable_type") {
		t.Errorf("expected plannable_type in JSON output, got: %s", output)
	}
}

func TestPlannerUpdateNoteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPlannerCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"planner", "update-note", "--note-id", "7001", "--title", "Updated Note", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "7001") {
		t.Errorf("expected note ID in JSON output, got: %s", output)
	}
}

