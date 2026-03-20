package linkedin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:1001") {
		t.Errorf("expected post URN in output, got: %s", out)
	}
	if !containsStr(out, "Hello LinkedIn world!") {
		t.Errorf("expected post text in output, got: %s", out)
	}
}

func TestPostsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON 'urn' field in output, got: %s", out)
	}
	if !containsStr(out, "urn:li:activity:1001") {
		t.Errorf("expected post URN in JSON output, got: %s", out)
	}
	if !containsStr(out, `"like_count"`) {
		t.Errorf("expected 'like_count' field in JSON output, got: %s", out)
	}
}

func TestPostsList_WithUsername(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "list", "--username", "testuser"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:1001") {
		t.Errorf("expected post in output with username filter, got: %s", out)
	}
}

func TestPostsList_Alias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"post", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:1001") {
		t.Errorf("expected post in output via alias, got: %s", out)
	}
}

func TestPostsList_Empty(t *testing.T) {
	server := newEmptyPostsServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "list"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No posts found.") {
		t.Errorf("expected 'No posts found.' in output, got: %s", out)
	}
}

func TestPostsGet_MissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"posts", "get"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --post-urn is missing")
	}
}

func TestPostsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "get", "--post-urn", "urn:li:activity:1001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:1001") {
		t.Errorf("expected post URN in output, got: %s", out)
	}
	if !containsStr(out, "Hello LinkedIn world!") {
		t.Errorf("expected post text in output, got: %s", out)
	}
}

func TestPostsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "get", "--post-urn", "urn:li:activity:1001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON 'urn' field in output, got: %s", out)
	}
	if !containsStr(out, "Hello LinkedIn world!") {
		t.Errorf("expected post text in JSON output, got: %s", out)
	}
}

func TestPostsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text", "Test post content", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestPostsCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text", "Test post content", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Test post content") {
		t.Errorf("expected post text in dry-run JSON output, got: %s", out)
	}
}

func TestPostsCreate_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text", "Hello LinkedIn!"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:activity:9999") {
		t.Errorf("expected created post URN in output, got: %s", out)
	}
}

func TestPostsCreate_Success_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "create", "--text", "Hello LinkedIn!", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON 'urn' field in output, got: %s", out)
	}
	if !containsStr(out, "urn:li:activity:9999") {
		t.Errorf("expected created post URN in JSON output, got: %s", out)
	}
}

func TestPostsCreate_MissingText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"posts", "create"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --text is missing")
	}
}

func TestPostsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "delete", "--post-urn", "urn:li:activity:1001", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestPostsDelete_RequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"posts", "delete", "--post-urn", "urn:li:activity:1001"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "irreversible") {
		t.Errorf("expected 'irreversible' in error message, got: %s", err.Error())
	}
}

func TestPostsDelete_WithConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "delete", "--post-urn", "urn:li:activity:1001", "--confirm"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestPostsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "delete", "--post-urn", "urn:li:activity:1001", "--confirm", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in JSON output, got: %s", out)
	}
}

func TestPostsReactions_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "reactions", "--post-urn", "urn:li:activity:1001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "LIKE") {
		t.Errorf("expected 'LIKE' reaction type in output, got: %s", out)
	}
	if !containsStr(out, "Alice Smith") {
		t.Errorf("expected actor name in output, got: %s", out)
	}
}

func TestPostsReactions_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "reactions", "--post-urn", "urn:li:activity:1001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"reaction_type"`) {
		t.Errorf("expected 'reaction_type' field in JSON output, got: %s", out)
	}
	if !containsStr(out, "LIKE") {
		t.Errorf("expected 'LIKE' in JSON output, got: %s", out)
	}
}

func TestPostsReactions_Empty(t *testing.T) {
	server := newEmptyPostsServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "reactions", "--post-urn", "urn:li:activity:1001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No reactions found.") {
		t.Errorf("expected 'No reactions found.' in output, got: %s", out)
	}
}

func TestPostsReact_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "react", "--post-urn", "urn:li:activity:1001", "--type", "LIKE", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestPostsReact_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "react", "--post-urn", "urn:li:activity:1001", "--type", "LIKE"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Reacted") {
		t.Errorf("expected 'Reacted' in output, got: %s", out)
	}
}

func TestPostsReact_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"posts", "react", "--post-urn", "urn:li:activity:1001", "--type", "CELEBRATE", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "reacted") {
		t.Errorf("expected 'reacted' in JSON output, got: %s", out)
	}
}

func TestPostsReact_MissingFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPostsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"posts", "react"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when required flags are missing")
	}
}

// newEmptyPostsServer returns a server that always returns empty element arrays.
func newEmptyPostsServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"elements":[],"paging":{"start":0,"count":10,"total":0}}`))
	}))
}
