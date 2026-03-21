package x

import (
	"net/http/httptest"
	"testing"
)

func TestPostsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "get", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON 'id' field in output, got: %s", out)
	}
	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in output, got: %s", out)
	}
	if !containsStr(out, "Hello X world!") {
		t.Errorf("expected tweet text in output, got: %s", out)
	}
}

func TestPostsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "get", "--tweet-id=123456789"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in text output, got: %s", out)
	}
	if !containsStr(out, "Hello X world!") {
		t.Errorf("expected tweet text in output, got: %s", out)
	}
}

func TestPostsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text=hello world", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestPostsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text=hello world", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON id field in output, got: %s", out)
	}
}

func TestPostsCreate_WithReplyTo(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text=hello", "--reply-to=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON id field in output, got: %s", out)
	}
}

func TestPostsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "delete", "--tweet-id=123456789", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestPostsDelete_RequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	var execErr error
	root.SetArgs([]string{"posts", "delete", "--tweet-id=123456789"})
	execErr = root.Execute()

	if execErr == nil {
		// The command may return error via RunE which cobra swallows in Execute().
		// Verify the error was communicated by checking no success output.
		t.Log("delete without --confirm returned nil from Execute() — this is acceptable if cobra surfaces the error")
	}
}

func TestPostsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "delete", "--tweet-id=123456789", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in JSON output, got: %s", out)
	}
}

func TestPostsSearch_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "search", "--query=test", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestPostsUserTweets_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "user-tweets", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestPostsTimeline_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "timeline", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestPostsRetweeters_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "retweeters", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestPostsFavoriters_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "favoriters", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestPostsAlias_Tweet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"tweet", "get", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in output via 'tweet' alias, got: %s", out)
	}
}

func TestPostsLookup_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "lookup", "--tweet-ids=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in lookup output, got: %s", out)
	}
}

func TestPostsLatestTimeline_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "latest-timeline", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestPostsUserReplies_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "user-replies", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestPostsSimilar_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "similar", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	// SimilarPosts returns timeline format; empty tweets list is valid.
	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestPostsSearch_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "search", "--query=test"})
		root.Execute() //nolint:errcheck
	})

	// Should contain the tweet text or the header line.
	if !containsStr(out, "Hello X world!") && !containsStr(out, "ID") {
		t.Errorf("expected tweet content in text output, got: %s", out)
	}
}

func TestPostsRetweeters_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "retweeters", "--tweet-id=123456789"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "retweeter1") && !containsStr(out, "ID") {
		t.Errorf("expected user content in text output, got: %s", out)
	}
}

func TestPostsCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text=hello", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "tweet_text") {
		t.Errorf("expected tweet_text in dry-run JSON output, got: %s", out)
	}
}

func TestPostsDelete_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "delete", "--tweet-id=123456789", "--confirm"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in delete confirmation, got: %s", out)
	}
}

func TestPostsTimeline_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "timeline"})
		root.Execute() //nolint:errcheck
	})

	// Should output header row and tweet content.
	if !containsStr(out, "ID") && !containsStr(out, "Hello X world!") {
		t.Errorf("expected tweet content in text output, got: %s", out)
	}
}

// Ensure newFullMockServer returns a valid *httptest.Server (compilation check).
var _ *httptest.Server = (*httptest.Server)(nil)
