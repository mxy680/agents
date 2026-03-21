package x

import (
	"testing"
)

func TestBlocksBlock_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"blocks", "block", "--user-id=222", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "blocked") {
		t.Errorf("expected 'blocked' in JSON output, got: %s", out)
	}
}

func TestBlocksBlock_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"blocks", "block", "--user-id=222"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "222") {
		t.Errorf("expected user ID in text output, got: %s", out)
	}
}

func TestBlocksBlock_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"blocks", "block", "--user-id=222", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBlocksBlock_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"blocks", "block", "--user-id=222", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "user_id") {
		t.Errorf("expected user_id in dry-run JSON output, got: %s", out)
	}
}

func TestBlocksUnblock_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"blocks", "unblock", "--user-id=222", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "unblocked") {
		t.Errorf("expected 'unblocked' in JSON output, got: %s", out)
	}
}

func TestBlocksUnblock_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"blocks", "unblock", "--user-id=222"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "222") {
		t.Errorf("expected user ID in text output, got: %s", out)
	}
}

func TestBlocksUnblock_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"blocks", "unblock", "--user-id=222", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestBlocksAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newBlocksCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"block", "block", "--user-id=222", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field via 'block' alias, got: %s", out)
	}
}
