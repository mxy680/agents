package x

import (
	"testing"
)

func TestRetweetsRetweet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"retweets", "retweet", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "retweeted") {
		t.Errorf("expected 'retweeted' in output, got: %s", out)
	}
}

func TestRetweetsRetweet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"retweets", "retweet", "--tweet-id=123456789"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "123456789") {
		t.Errorf("expected tweet ID in text output, got: %s", out)
	}
}

func TestRetweetsRetweet_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"retweets", "retweet", "--tweet-id=123456789", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestRetweetsRetweet_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"retweets", "retweet", "--tweet-id=123456789", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "tweet_id") {
		t.Errorf("expected tweet_id in dry-run JSON output, got: %s", out)
	}
}

func TestRetweetsUndo_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"retweets", "undo", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "retweet_undone") {
		t.Errorf("expected 'retweet_undone' in output, got: %s", out)
	}
}

func TestRetweetsUndo_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"retweets", "undo", "--tweet-id=123456789", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestRetweetsAlias_RT(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"rt", "retweet", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "retweeted") {
		t.Errorf("expected 'retweeted' via 'rt' alias, got: %s", out)
	}
}

func TestRetweetsAlias_Retweet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newRetweetsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"retweet", "undo", "--tweet-id=123456789", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "retweet_undone") {
		t.Errorf("expected 'retweet_undone' via 'retweet' alias, got: %s", out)
	}
}
