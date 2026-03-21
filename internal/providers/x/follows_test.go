package x

import (
	"testing"
)

func TestFollowsFollowers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "followers", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestFollowsFollowers_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "followers", "--user-id=999"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "follower1") && !containsStr(out, "ID") {
		t.Errorf("expected user content in text output, got: %s", out)
	}
}

func TestFollowsFollowing_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "following", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestFollowsFollowing_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "following", "--user-id=999"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "follower1") && !containsStr(out, "ID") {
		t.Errorf("expected user content in text output, got: %s", out)
	}
}

func TestFollowsVerifiedFollowers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "verified-followers", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestFollowsFollowersYouKnow_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "followers-you-know", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestFollowsFollow_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "follow", "--user-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "followed") {
		t.Errorf("expected 'followed' in JSON output, got: %s", out)
	}
}

func TestFollowsFollow_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "follow", "--user-id=111"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "111") {
		t.Errorf("expected user ID in text output, got: %s", out)
	}
}

func TestFollowsFollow_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "follow", "--user-id=111", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestFollowsFollow_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "follow", "--user-id=111", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "user_id") {
		t.Errorf("expected user_id in dry-run JSON output, got: %s", out)
	}
}

func TestFollowsUnfollow_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "unfollow", "--user-id=111", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON status field in output, got: %s", out)
	}
	if !containsStr(out, "unfollowed") {
		t.Errorf("expected 'unfollowed' in JSON output, got: %s", out)
	}
}

func TestFollowsUnfollow_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "unfollow", "--user-id=111"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "111") {
		t.Errorf("expected user ID in text output, got: %s", out)
	}
}

func TestFollowsUnfollow_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "unfollow", "--user-id=111", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestFollowsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follow", "followers", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field via 'follow' alias, got: %s", out)
	}
}

func TestFollowsFollowers_WithCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFollowsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follows", "followers", "--user-id=999", "--cursor=abc123", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output with cursor, got: %s", out)
	}
}
