package x

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "get", "--list-id=list-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON id field in output, got: %s", out)
	}
	if !containsStr(out, "list-001") {
		t.Errorf("expected list ID in output, got: %s", out)
	}
}

func TestListsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "get", "--list-id=list-001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "My Test List") && !containsStr(out, "Name:") {
		t.Errorf("expected list name in text output, got: %s", out)
	}
}

func TestListsOwned_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "owned", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"lists"`) {
		t.Errorf("expected JSON lists field in output, got: %s", out)
	}
}

func TestListsOwned_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "owned"})
		root.Execute() //nolint:errcheck
	})

	// Output may be "No lists found." or contain list data.
	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestListsOwned_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "owned", "--cursor=abc123", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"lists"`) {
		t.Errorf("expected JSON lists field in output, got: %s", out)
	}
}

func TestListsSearch_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "search", "--query=test", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"lists"`) {
		t.Errorf("expected JSON lists field in output, got: %s", out)
	}
}

func TestListsTweets_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "tweets", "--list-id=list-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in output, got: %s", out)
	}
}

func TestListsTweets_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "tweets", "--list-id=list-001", "--cursor=abc", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestListsMembers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "members", "--list-id=list-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
	if !containsStr(out, "listmember") {
		t.Errorf("expected member username in output, got: %s", out)
	}
}

func TestListsSubscribers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "subscribers", "--list-id=list-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestListsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "create", "--name=My Test List", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "list-001") && !containsStr(out, `"name"`) {
		t.Errorf("expected list data in output, got: %s", out)
	}
}

func TestListsCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "create", "--name=My Test List"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "created") {
		t.Errorf("expected 'created' in output, got: %s", out)
	}
}

func TestListsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "create", "--name=Test List", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestListsCreate_Private(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "create", "--name=Private List", "--private", "--json"})
		root.Execute() //nolint:errcheck
	})

	// As long as there's output, the command ran.
	if out == "" {
		t.Error("expected some output, got empty string")
	}
}

func TestListsUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "update", "--list-id=list-001", "--name=Updated Name", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "updated") {
		t.Errorf("expected 'updated' in output, got: %s", out)
	}
}

func TestListsUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "update", "--list-id=list-001", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestListsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "delete", "--list-id=list-001", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestListsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "delete", "--list-id=list-001", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestListsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"lists", "delete", "--list-id=list-001"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --confirm not provided, got nil")
	}
}

func TestListsAddMember_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "add-member", "--list-id=list-001", "--user-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "added") {
		t.Errorf("expected 'added' in output, got: %s", out)
	}
}

func TestListsAddMember_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "add-member", "--list-id=list-001", "--user-id=111", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestListsRemoveMember_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "remove-member", "--list-id=list-001", "--user-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "removed") {
		t.Errorf("expected 'removed' in output, got: %s", out)
	}
}

func TestListsRemoveMember_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "remove-member", "--list-id=list-001", "--user-id=111", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestListsSetBanner_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "set-banner", "--list-id=list-001", "--path=/tmp/banner.png", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestListsSetBanner_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	// Create a temporary image file for the test.
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "banner.png")
	// Write minimal valid PNG bytes.
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := os.WriteFile(imgPath, pngHeader, 0600); err != nil {
		t.Fatalf("create temp image: %v", err)
	}

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "set-banner", "--list-id=list-001", "--path=" + imgPath, "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "banner_set") {
		t.Errorf("expected 'banner_set' in output, got: %s", out)
	}
}

func TestListsRemoveBanner_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "remove-banner", "--list-id=list-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "banner_removed") {
		t.Errorf("expected 'banner_removed' in output, got: %s", out)
	}
}

func TestListsRemoveBanner_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"lists", "remove-banner", "--list-id=list-001", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestListsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newListsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"list", "get", "--list-id=list-001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "list-001") {
		t.Errorf("expected list ID in output via alias, got: %s", out)
	}
}
