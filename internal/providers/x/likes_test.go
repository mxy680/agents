package x

import (
	"testing"
)

func TestLikesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "list", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in output, got: %s", out)
	}
}

func TestLikesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "list", "--user-id=999"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Hello X world!") && !containsStr(out, "ID") {
		t.Errorf("expected tweet content in text output, got: %s", out)
	}
}

func TestLikesList_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "list", "--user-id=999", "--cursor=abc123", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestLikesLike_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "like", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "liked") {
		t.Errorf("expected 'liked' in output, got: %s", out)
	}
}

func TestLikesLike_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "like", "--tweet-id=123456789"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in output, got: %s", out)
	}
}

func TestLikesLike_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "like", "--tweet-id=123456789", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestLikesLike_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "like", "--tweet-id=123456789", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "tweet_id") {
		t.Errorf("expected tweet_id in dry-run JSON output, got: %s", out)
	}
}

func TestLikesUnlike_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "unlike", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "unliked") {
		t.Errorf("expected 'unliked' in output, got: %s", out)
	}
}

func TestLikesUnlike_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"likes", "unlike", "--tweet-id=123456789", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestLikesAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newLikesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"like", "list", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field via 'like' alias, got: %s", out)
	}
}
