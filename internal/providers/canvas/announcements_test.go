package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnnouncementsListText(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"announcements", "list", "--course-ids", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Welcome Announcement") {
		t.Errorf("expected first announcement title, got: %s", output)
	}
	if !strings.Contains(output, "Midterm Reminder") {
		t.Errorf("expected second announcement title, got: %s", output)
	}
}

func TestAnnouncementsListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"announcements", "list", "--course-ids", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("JSON output should contain title field, got: %s", output)
	}
	if !strings.Contains(output, "Welcome Announcement") {
		t.Errorf("JSON output should contain announcement title, got: %s", output)
	}
}

func TestAnnouncementsListMissingCourseIDs(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"announcements", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-ids is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-ids") {
		t.Errorf("error should mention --course-ids, got: %v", execErr)
	}
}

func TestAnnouncementsGetText(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"announcements", "get",
			"--course-id", "101",
			"--announcement-id", "301",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Welcome Announcement") {
		t.Errorf("expected announcement title, got: %s", output)
	}
	if !strings.Contains(output, "Instructor") {
		t.Errorf("expected author name, got: %s", output)
	}
	if !strings.Contains(output, "Welcome to the course!") {
		t.Errorf("expected announcement message, got: %s", output)
	}
}

func TestAnnouncementsCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"announcements", "create",
			"--course-id", "101",
			"--title", "Important Update",
			"--message", "Please read this.",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "Important Update") {
		t.Errorf("expected announcement title in dry-run output, got: %s", output)
	}
}

func TestAnnouncementsDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{
			"announcements", "delete",
			"--course-id", "101",
			"--announcement-id", "301",
		})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestAnnouncementsDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"announcements", "delete",
			"--course-id", "101",
			"--announcement-id", "301",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "301") {
		t.Errorf("expected announcement ID in output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestAnnouncementsCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"announcements", "create",
			"--course-id", "101",
			"--title", "New Announcement",
			"--message", "Hello class",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "202") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestAnnouncementsUpdateLive(t *testing.T) {
	mux := http.NewServeMux()
	withAnnouncementsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newAnnouncementsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"announcements", "update",
			"--course-id", "101",
			"--announcement-id", "301",
			"--title", "Updated Title",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "301") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output with ID 301, got: %s", output)
	}
}
