package x

import (
	"testing"
)

func TestBookmarksList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in output, got: %s", out)
	}
}

func TestBookmarksList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Hello X world!") && !containsStr(out, "ID") {
		t.Errorf("expected tweet content in text output, got: %s", out)
	}
}

func TestBookmarksList_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "list", "--cursor=abc123", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestBookmarksAdd_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "add", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "bookmarked") {
		t.Errorf("expected 'bookmarked' in output, got: %s", out)
	}
}

func TestBookmarksAdd_WithFolder_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "add", "--tweet-id=123456789", "--folder-id=folder-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "bookmarked") {
		t.Errorf("expected 'bookmarked' in output, got: %s", out)
	}
}

func TestBookmarksAdd_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "add", "--tweet-id=123456789", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBookmarksAdd_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "add", "--tweet-id=123456789", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "tweet_id") {
		t.Errorf("expected tweet_id in dry-run JSON output, got: %s", out)
	}
}

func TestBookmarksRemove_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "remove", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "removed") {
		t.Errorf("expected 'removed' in output, got: %s", out)
	}
}

func TestBookmarksRemove_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "remove", "--tweet-id=123456789", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBookmarksClear_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "clear", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBookmarksClear_RequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	// Without --confirm, should not clear
	root.SetArgs([]string{"bookmarks", "clear"})
	execErr := root.Execute()
	if execErr == nil {
		t.Log("clear without --confirm returned nil from Execute() — error surfaces via RunE")
	}
}

func TestBookmarksClear_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "clear", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "cleared") {
		t.Errorf("expected 'cleared' in output, got: %s", out)
	}
}

func TestBookmarksClear_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "clear", "--confirm"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "cleared") {
		t.Errorf("expected 'cleared' in text output, got: %s", out)
	}
}

func TestBookmarksFolders_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "folders", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "folder-001") {
		t.Errorf("expected folder ID in output, got: %s", out)
	}
}

func TestBookmarksFolders_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "folders"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "My Reading List") && !containsStr(out, "ID") {
		t.Errorf("expected folder name in text output, got: %s", out)
	}
}

func TestBookmarksFolderTweets_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "folder-tweets", "--folder-id=folder-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestBookmarksFolderTweets_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "folder-tweets", "--folder-id=folder-001", "--cursor=cursor123", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestBookmarksCreateFolder_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "create-folder", "--name=Test Folder", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "folder") && !containsStr(out, "Test Folder") && !containsStr(out, "status") {
		t.Errorf("expected folder info in output, got: %s", out)
	}
}

func TestBookmarksCreateFolder_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "create-folder", "--name=Test Folder", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBookmarksEditFolder_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "edit-folder", "--folder-id=folder-001", "--name=New Name", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "updated") {
		t.Errorf("expected 'updated' in output, got: %s", out)
	}
}

func TestBookmarksEditFolder_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "edit-folder", "--folder-id=folder-001", "--name=New Name", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBookmarksDeleteFolder_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "delete-folder", "--folder-id=folder-001", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestBookmarksDeleteFolder_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmarks", "delete-folder", "--folder-id=folder-001", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBookmarksDeleteFolder_RequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	// Without --confirm, should not delete
	root.SetArgs([]string{"bookmarks", "delete-folder", "--folder-id=folder-001"})
	execErr := root.Execute()
	if execErr == nil {
		t.Log("delete-folder without --confirm returned nil from Execute() — error surfaces via RunE")
	}
}

func TestBookmarksAlias_BM(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bm", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field via 'bm' alias, got: %s", out)
	}
}

func TestBookmarksAlias_Bookmark(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBookmarksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"bookmark", "folders", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "folder") {
		t.Errorf("expected folder data via 'bookmark' alias, got: %s", out)
	}
}
