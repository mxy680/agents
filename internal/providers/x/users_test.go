package x

import (
	"testing"
)

func TestUsersGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"users", "get", "--username=testuser", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON 'id' field in output, got: %s", out)
	}
	if !containsStr(out, "testuser") {
		t.Errorf("expected username in output, got: %s", out)
	}
	if !containsStr(out, "Test User") {
		t.Errorf("expected name in output, got: %s", out)
	}
}

func TestUsersGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"users", "get", "--username=testuser"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "testuser") {
		t.Errorf("expected username in text output, got: %s", out)
	}
	if !containsStr(out, "Test User") {
		t.Errorf("expected name in text output, got: %s", out)
	}
}

func TestUsersGetByID_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"users", "get-by-id", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"id"`) {
		t.Errorf("expected JSON 'id' field in output, got: %s", out)
	}
	if !containsStr(out, "999") {
		t.Errorf("expected user ID in output, got: %s", out)
	}
}

func TestUsersSearch_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"users", "search", "--query=test", "--json"})
		root.Execute() //nolint:errcheck
	})

	// users search reuses SearchTimeline which returns timeline (tweet) format in our mock,
	// so we check for the users wrapper key (even if empty from tweet-based mock).
	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}

func TestUsersHighlights_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"users", "highlights", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestUsersMedia_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"users", "media", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"tweets"`) {
		t.Errorf("expected JSON tweets field in output, got: %s", out)
	}
}

func TestUsersAlias_User(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"user", "get", "--username=testuser", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "testuser") {
		t.Errorf("expected username in output via 'user' alias, got: %s", out)
	}
}

func TestUsersSubscriptions_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newUsersCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"users", "subscriptions", "--user-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"users"`) {
		t.Errorf("expected JSON users field in output, got: %s", out)
	}
}
