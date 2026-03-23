package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiscussionsListText(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"discussions", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Discussion Topic One") {
		t.Errorf("expected first topic title, got: %s", output)
	}
	if !strings.Contains(output, "Discussion Topic Two") {
		t.Errorf("expected second topic title, got: %s", output)
	}
}

func TestDiscussionsListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"discussions", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("JSON output should contain title field, got: %s", output)
	}
	if !strings.Contains(output, "Discussion Topic One") {
		t.Errorf("JSON output should contain first topic, got: %s", output)
	}
}

func TestDiscussionsListMissingCourseID(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"discussions", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestDiscussionsGetText(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"discussions", "get", "--course-id", "101", "--topic-id", "201"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Discussion Topic One") {
		t.Errorf("expected topic title, got: %s", output)
	}
	if !strings.Contains(output, "Instructor") {
		t.Errorf("expected author name, got: %s", output)
	}
}

func TestDiscussionsGetJSON(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"discussions", "get", "--course-id", "101", "--topic-id", "201", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("JSON output should contain title field, got: %s", output)
	}
	if !strings.Contains(output, "Discussion Topic One") {
		t.Errorf("JSON output should contain topic title, got: %s", output)
	}
}

func TestDiscussionsCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "create",
			"--course-id", "101",
			"--title", "New Topic",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "New Topic") {
		t.Errorf("expected topic title in dry-run output, got: %s", output)
	}
}

func TestDiscussionsDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"discussions", "delete", "--course-id", "101", "--topic-id", "201"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestDiscussionsDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "delete",
			"--course-id", "101",
			"--topic-id", "201",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "201") {
		t.Errorf("expected topic ID in output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestDiscussionsEntriesText(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "entries",
			"--course-id", "101",
			"--topic-id", "201",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Test User") {
		t.Errorf("expected user name in output, got: %s", output)
	}
	if !strings.Contains(output, "First entry message") {
		t.Errorf("expected entry message in output, got: %s", output)
	}
}

func TestDiscussionsReplyDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "reply",
			"--course-id", "101",
			"--topic-id", "201",
			"--message", "My reply text",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestDiscussionsMarkRead(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "mark-read",
			"--course-id", "101",
			"--topic-id", "201",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "201") {
		t.Errorf("expected topic ID in output, got: %s", output)
	}
	if !strings.Contains(output, "read") {
		t.Errorf("expected 'read' in output, got: %s", output)
	}
}

func TestDiscussionsMarkReadDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "mark-read",
			"--course-id", "101",
			"--topic-id", "201",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestDiscussionsCreateLive(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "create",
			"--course-id", "101",
			"--title", "New Discussion",
			"--message", "Let's discuss.",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "202") && !strings.Contains(output, "New Discussion Topic") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestDiscussionsUpdateLive(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "update",
			"--course-id", "101",
			"--topic-id", "201",
			"--title", "Updated Topic",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "201") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output, got: %s", output)
	}
}

func TestDiscussionsReplyLive(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "reply",
			"--course-id", "101",
			"--topic-id", "201",
			"--message", "This is my reply",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "reply") && !strings.Contains(output, "3001") {
		t.Errorf("expected reply output, got: %s", output)
	}
}

func TestDiscussionsMarkReadLive(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "mark-read",
			"--course-id", "101",
			"--topic-id", "201",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "201") && !strings.Contains(output, "read") {
		t.Errorf("expected mark-read output, got: %s", output)
	}
}

func TestDiscussionsEntriesJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "entries",
			"--course-id", "101",
			"--topic-id", "201",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "message") {
		t.Errorf("expected JSON entries output with message field, got: %s", output)
	}
}

func TestDiscussionsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "update",
			"--course-id", "101",
			"--topic-id", "201",
			"--title", "Updated Topic",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "201") {
		t.Errorf("expected discussion ID in JSON output, got: %s", output)
	}
}

func TestDiscussionsUpdateWithOptionalFlags(t *testing.T) {
	mux := http.NewServeMux()
	withDiscussionsMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newDiscussionsCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"discussions", "update",
			"--course-id", "101",
			"--topic-id", "201",
			"--title", "Updated Topic",
			"--message", "New body",
			"--pinned",
			"--locked",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "201") {
		t.Errorf("expected discussion ID in output, got: %s", output)
	}
}
