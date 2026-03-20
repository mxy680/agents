package linkedin

import (
	"testing"
)

func TestFeedList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"feed", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:2001") {
		t.Errorf("expected feed post URN in output, got: %s", out)
	}
	if !containsStr(out, "Interesting feed post") {
		t.Errorf("expected feed post text in output, got: %s", out)
	}
}

func TestFeedList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"feed", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON 'urn' field in output, got: %s", out)
	}
	if !containsStr(out, "urn:li:activity:2001") {
		t.Errorf("expected feed post URN in JSON output, got: %s", out)
	}
	if !containsStr(out, `"like_count"`) {
		t.Errorf("expected 'like_count' field in JSON output, got: %s", out)
	}
}

func TestFeedList_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"feed", "list", "--limit", "5", "--cursor", "10"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:2001") {
		t.Errorf("expected feed post in output with cursor, got: %s", out)
	}
}

func TestFeedList_Empty(t *testing.T) {
	server := newEmptyPostsServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"feed", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No posts found.") {
		t.Errorf("expected 'No posts found.' in output, got: %s", out)
	}
}

func TestFeedHashtag_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"feed", "hashtag", "--tag", "golang"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:2001") {
		t.Errorf("expected feed post in hashtag output, got: %s", out)
	}
}

func TestFeedHashtag_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"feed", "hashtag", "--tag", "golang", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON 'urn' field in hashtag output, got: %s", out)
	}
	if !containsStr(out, "urn:li:activity:2001") {
		t.Errorf("expected feed post URN in hashtag JSON output, got: %s", out)
	}
}

func TestFeedHashtag_MissingTag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"feed", "hashtag"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --tag is missing")
	}
}

func TestFeedHashtag_Empty(t *testing.T) {
	server := newEmptyPostsServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"feed", "hashtag", "--tag", "obscuretag"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No posts found.") {
		t.Errorf("expected 'No posts found.' in hashtag output, got: %s", out)
	}
}
