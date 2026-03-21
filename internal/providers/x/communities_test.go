package x

import (
	"testing"
)

func TestCommunitiesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "get", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "111") && !containsStr(out, "community") {
		t.Errorf("expected community ID or data in output, got: %s", out)
	}
}

func TestCommunitiesSearch_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "search", "--query=golang", "--json"})
		root.Execute() //nolint:errcheck
	})

	// Either returns community data or empty list — both are valid.
	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestCommunityTweets_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "tweets", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestCommunityMedia_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "media", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestCommunityMembers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "members", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestCommunityModerators_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "moderators", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestCommunitiesTimeline_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "timeline", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestCommunityJoin_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "join", "--community-id=111", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestCommunityJoin_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "join", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "joined") {
		t.Errorf("expected 'joined' in output, got: %s", out)
	}
}

func TestCommunityLeave_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "leave", "--community-id=111", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestCommunityLeave_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "leave", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "left") {
		t.Errorf("expected 'left' in output, got: %s", out)
	}
}

func TestCommunityRequestJoin_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "request-join", "--community-id=111", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestCommunityRequestJoin_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "request-join", "--community-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "requested") {
		t.Errorf("expected 'requested' in output, got: %s", out)
	}
}

func TestCommunitySearchTweets_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"communities", "search-tweets", "--community-id=111", "--query=test", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestCommunitiesAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommunitiesCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"community", "timeline", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field via 'community' alias, got: %s", out)
	}
}
