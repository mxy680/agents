package canvas

import (
	"strings"
	"testing"
)

func TestBookmarksListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "list"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("expected bookmark name 'Intro to CS', got: %s", output)
	}
	if !strings.Contains(output, "Data Structures") {
		t.Errorf("expected bookmark name 'Data Structures', got: %s", output)
	}
}

func TestBookmarksGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "get", "--bookmark-id", "8001"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("expected bookmark name, got: %s", output)
	}
	if !strings.Contains(output, "canvas.edu") {
		t.Errorf("expected URL in output, got: %s", output)
	}
}

func TestBookmarksGetMissingID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "get"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --bookmark-id is missing")
	}
	if !strings.Contains(execErr.Error(), "--bookmark-id") {
		t.Errorf("error should mention --bookmark-id, got: %v", execErr)
	}
}

func TestBookmarksCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "create", "--name", "My Bookmark", "--url", "https://canvas.edu/courses/101", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
}

func TestBookmarksDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "delete", "--bookmark-id", "8001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm is missing")
	}
	if !strings.Contains(execErr.Error(), "--confirm") {
		t.Errorf("error should mention --confirm, got: %v", execErr)
	}
}

func TestBookmarksCreateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "create", "--name", "My Bookmark", "--url", "https://canvas.edu/courses/101"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "created") && !strings.Contains(output, "New Bookmark") && !strings.Contains(output, "8001") {
		t.Errorf("expected creation output, got: %s", output)
	}
}

func TestBookmarksUpdateLive(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "update", "--bookmark-id", "8001", "--name", "Updated Bookmark"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "8001") || !strings.Contains(output, "updated") {
		t.Errorf("expected update output, got: %s", output)
	}
}

func TestBookmarksDeleteConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "delete", "--bookmark-id", "8001", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "8001") || !strings.Contains(output, "deleted") {
		t.Errorf("expected deletion output, got: %s", output)
	}
}

func TestBookmarksListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("expected bookmark name in JSON output, got: %s", output)
	}
}

func TestBookmarksUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "update", "--bookmark-id", "8001", "--name", "Updated Bookmark", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "8001") {
		t.Errorf("expected bookmark ID in JSON output, got: %s", output)
	}
}

func TestBookmarksGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "get", "--bookmark-id", "8001", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("expected bookmark name in JSON output, got: %s", output)
	}
}
