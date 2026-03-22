package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalendarListText(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"calendar", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Midterm Exam") {
		t.Errorf("expected event title in output, got: %s", output)
	}
	if !strings.Contains(output, "Office Hours") {
		t.Errorf("expected second event title in output, got: %s", output)
	}
	if !strings.Contains(output, "course_101") {
		t.Errorf("expected context code in output, got: %s", output)
	}
}

func TestCalendarListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"calendar", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("expected title field in JSON, got: %s", output)
	}
	if !strings.Contains(output, "Midterm Exam") {
		t.Errorf("expected event title in JSON output, got: %s", output)
	}
	if !strings.Contains(output, `"context_code"`) {
		t.Errorf("expected context_code field in JSON, got: %s", output)
	}
}

func TestCalendarGetText(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"calendar", "get", "--event-id", "901"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Midterm Exam") {
		t.Errorf("expected event title in output, got: %s", output)
	}
	if !strings.Contains(output, "901") {
		t.Errorf("expected event ID in output, got: %s", output)
	}
	if !strings.Contains(output, "course_101") {
		t.Errorf("expected context code in output, got: %s", output)
	}
	if !strings.Contains(output, "Room 204") {
		t.Errorf("expected location in output, got: %s", output)
	}
	if !strings.Contains(output, "active") {
		t.Errorf("expected workflow state in output, got: %s", output)
	}
}

func TestCalendarGetMissingID(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"calendar", "get"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --event-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--event-id") {
		t.Errorf("error should mention --event-id, got: %v", execErr)
	}
}

func TestCalendarCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"calendar", "create",
			"--context-code", "course_101",
			"--title", "Study Session",
			"--start-at", "2026-04-01T10:00:00Z",
			"--end-at", "2026-04-01T12:00:00Z",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
	if !strings.Contains(output, "Study Session") {
		t.Errorf("expected event title in dry-run output, got: %s", output)
	}
}

func TestCalendarCreateSuccess(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"calendar", "create",
			"--context-code", "course_101",
			"--title", "Office Hours",
			"--start-at", "2026-02-10T14:00:00Z",
			"--end-at", "2026-02-10T15:00:00Z",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Office Hours") {
		t.Errorf("expected event title in output, got: %s", output)
	}
}

func TestCalendarDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"calendar", "delete", "--event-id", "901"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestCalendarDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withCalendarMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newCalendarCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"calendar", "delete", "--event-id", "901", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "901") {
		t.Errorf("expected event ID in deletion output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}
