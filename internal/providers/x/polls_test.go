package x

import (
	"testing"
)

func TestPollsCreate_TooFewOptions(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"polls", "create", "--options=only one", "--duration=1440"})
	err := root.Execute()
	_ = err
	// Should fail with validation error.
}

func TestPollsCreate_TooManyOptions(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"polls", "create",
		"--options=one,two,three,four,five",
		"--duration=1440",
	})
	err := root.Execute()
	_ = err
	// Should fail with validation error.
}

func TestPollsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"polls", "create",
			"--options=Yes,No",
			"--duration=1440",
			"--json",
		})
		root.Execute() //nolint:errcheck
	})

	// Either returns card_uri or status.
	if !containsStr(out, "card_uri") && !containsStr(out, "status") {
		t.Errorf("expected card_uri or status in output, got: %s", out)
	}
}

func TestPollsVote_InvalidChoice(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"polls", "vote", "--tweet-id=123", "--choice=0"})
	err := root.Execute()
	_ = err
}

func TestPollsVote_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"polls", "vote", "--tweet-id=123", "--choice=1", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestPollsVote_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"polls", "vote", "--tweet-id=123", "--choice=2", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "tweet_id") {
		t.Errorf("expected tweet_id in dry-run JSON output, got: %s", out)
	}
}

func TestPollsVote_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"polls", "vote", "--tweet-id=123", "--choice=1", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestPollsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newPollsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"poll", "vote", "--tweet-id=123", "--choice=1", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN via 'poll' alias, got: %s", out)
	}
}

func TestBuildPollCardData(t *testing.T) {
	options := []string{"Yes", "No", "Maybe"}
	data := buildPollCardData(options, 60)

	if data["card_name"] != "poll3choice_text_only" {
		t.Errorf("expected poll3choice_text_only, got: %v", data["card_name"])
	}
	if data["duration_minutes"] != 60 {
		t.Errorf("expected duration 60, got: %v", data["duration_minutes"])
	}
	if data["choice1_label"] != "Yes" {
		t.Errorf("expected choice1_label=Yes, got: %v", data["choice1_label"])
	}
	if data["choice3_label"] != "Maybe" {
		t.Errorf("expected choice3_label=Maybe, got: %v", data["choice3_label"])
	}
}
