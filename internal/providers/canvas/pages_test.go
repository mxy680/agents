package canvas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPagesListText(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"pages", "list", "--course-id", "101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Test Page") {
		t.Errorf("expected first page title, got: %s", output)
	}
	if !strings.Contains(output, "Syllabus") {
		t.Errorf("expected second page title, got: %s", output)
	}
	if !strings.Contains(output, "published") {
		t.Errorf("expected published status in output, got: %s", output)
	}
}

func TestPagesListJSON(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"pages", "list", "--course-id", "101", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("JSON output should contain title field, got: %s", output)
	}
	if !strings.Contains(output, "Test Page") {
		t.Errorf("JSON output should contain page title, got: %s", output)
	}
	if !strings.Contains(output, `"url"`) {
		t.Errorf("JSON output should contain url field, got: %s", output)
	}
}

func TestPagesListMissingCourseID(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"pages", "list"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --course-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--course-id") {
		t.Errorf("error should mention --course-id, got: %v", execErr)
	}
}

func TestPagesGetText(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"pages", "get", "--course-id", "101", "--url", "test-page"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Test Page") {
		t.Errorf("expected page title, got: %s", output)
	}
	if !strings.Contains(output, "test-page") {
		t.Errorf("expected page URL slug, got: %s", output)
	}
	if !strings.Contains(output, "teachers") {
		t.Errorf("expected editing roles, got: %s", output)
	}
}

func TestPagesGetJSON(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"pages", "get", "--course-id", "101", "--url", "test-page", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"title"`) {
		t.Errorf("JSON output should contain title field, got: %s", output)
	}
	if !strings.Contains(output, "Test Page") {
		t.Errorf("JSON output should contain page title, got: %s", output)
	}
	if !strings.Contains(output, `"url"`) {
		t.Errorf("JSON output should contain url field, got: %s", output)
	}
}

func TestPagesCreateDryRun(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"pages", "create",
			"--course-id", "101",
			"--title", "My New Page",
			"--dry-run",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
	if !strings.Contains(output, "My New Page") {
		t.Errorf("expected page title in dry-run output, got: %s", output)
	}
}

func TestPagesDeleteNoConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"pages", "delete", "--course-id", "101", "--url", "test-page"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is absent")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestPagesDeleteConfirm(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"pages", "delete",
			"--course-id", "101",
			"--url", "test-page",
			"--confirm",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "test-page") {
		t.Errorf("expected page URL slug in output, got: %s", output)
	}
	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", output)
	}
}

func TestPagesRevisionsText(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"pages", "revisions",
			"--course-id", "101",
			"--url", "test-page",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Instructor") {
		t.Errorf("expected editor name in output, got: %s", output)
	}
	if !strings.Contains(output, "[latest]") {
		t.Errorf("expected [latest] marker for most recent revision, got: %s", output)
	}
}

func TestPagesCreateLive(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"pages", "create",
			"--course-id", "101",
			"--title", "New Page",
			"--body", "Page content here",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "new-page") && !strings.Contains(output, "New Page") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestPagesUpdateLive(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"pages", "update",
			"--course-id", "101",
			"--url", "test-page",
			"--title", "Updated Test Page",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "test-page") {
		t.Errorf("expected page URL in update output, got: %s", output)
	}
}

func TestPagesUpdateJSON(t *testing.T) {
	mux := http.NewServeMux()
	withPagesMock(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newPagesCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{
			"pages", "update",
			"--course-id", "101",
			"--url", "test-page",
			"--title", "Updated Test Page",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "test-page") {
		t.Errorf("expected page URL in JSON output, got: %s", output)
	}
}

