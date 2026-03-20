package linkedin

import (
	"testing"
)

func TestCommentsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "list", "--post-urn", "urn:li:activity:1001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Great post!") {
		t.Errorf("expected comment text in output, got: %s", out)
	}
	if !containsStr(out, "Jane Doe") {
		t.Errorf("expected author name in output, got: %s", out)
	}
}

func TestCommentsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "list", "--post-urn", "urn:li:activity:1001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON 'urn' field in output, got: %s", out)
	}
	if !containsStr(out, "Great post!") {
		t.Errorf("expected comment text in JSON output, got: %s", out)
	}
	if !containsStr(out, `"like_count"`) {
		t.Errorf("expected 'like_count' field in JSON output, got: %s", out)
	}
}

func TestCommentsList_Alias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comment", "list", "--post-urn", "urn:li:activity:1001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Great post!") {
		t.Errorf("expected comment text in output via alias, got: %s", out)
	}
}

func TestCommentsList_MissingPostURN(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"comments", "list"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --post-urn is missing")
	}
}

func TestCommentsList_Empty(t *testing.T) {
	server := newEmptyPostsServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "list", "--post-urn", "urn:li:activity:1001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "No comments found.") {
		t.Errorf("expected 'No comments found.' in output, got: %s", out)
	}
}

func TestCommentsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "create",
			"--post-urn", "urn:li:activity:1001",
			"--text", "Awesome post!",
			"--dry-run",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestCommentsCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "create",
			"--post-urn", "urn:li:activity:1001",
			"--text", "Awesome post!",
			"--dry-run", "--json",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Awesome post!") {
		t.Errorf("expected comment text in dry-run JSON output, got: %s", out)
	}
}

func TestCommentsCreate_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "create",
			"--post-urn", "urn:li:activity:1001",
			"--text", "Awesome post!",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "urn:li:comment:(activity:1001,6001)") {
		t.Errorf("expected created comment URN in output, got: %s", out)
	}
}

func TestCommentsCreate_WithReplyTo(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "create",
			"--post-urn", "urn:li:activity:1001",
			"--text", "Replying here",
			"--reply-to", "urn:li:comment:(activity:1001,5001)",
		})
		root.Execute() //nolint:errcheck
	})

	// Should still produce a valid URN (mock doesn't distinguish reply-to)
	if !containsStr(out, "Comment created") {
		t.Errorf("expected 'Comment created' in output, got: %s", out)
	}
}

func TestCommentsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "create",
			"--post-urn", "urn:li:activity:1001",
			"--text", "Awesome!",
			"--json",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON 'urn' field in output, got: %s", out)
	}
}

func TestCommentsCreate_MissingFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"comments", "create"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when required flags are missing")
	}
}

func TestCommentsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "delete",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
			"--dry-run",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestCommentsDelete_RequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"comments", "delete", "--comment-urn", "urn:li:comment:(activity:1001,5001)"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "irreversible") {
		t.Errorf("expected 'irreversible' in error message, got: %s", err.Error())
	}
}

func TestCommentsDelete_WithConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "delete",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
			"--confirm",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestCommentsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "delete",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
			"--confirm", "--json",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "deleted") {
		t.Errorf("expected 'deleted' in JSON output, got: %s", out)
	}
}

func TestCommentsLike_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "like",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
			"--dry-run",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestCommentsLike_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "like",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Liked") {
		t.Errorf("expected 'Liked' in output, got: %s", out)
	}
}

func TestCommentsLike_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "like",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
			"--json",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "liked") {
		t.Errorf("expected 'liked' in JSON output, got: %s", out)
	}
}

func TestCommentsUnlike_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "unlike",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
			"--dry-run",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestCommentsUnlike_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "unlike",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Unliked") {
		t.Errorf("expected 'Unliked' in output, got: %s", out)
	}
}

func TestCommentsUnlike_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newCommentsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"comments", "unlike",
			"--comment-urn", "urn:li:comment:(activity:1001,5001)",
			"--json",
		})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "unliked") {
		t.Errorf("expected 'unliked' in JSON output, got: %s", out)
	}
}
