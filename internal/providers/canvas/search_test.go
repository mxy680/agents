package canvas

import (
	"strings"
	"testing"
)

func TestSearchRecipientsText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"search", "recipients", "--search", "alice"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Alice Student") {
		t.Errorf("expected 'Alice Student' in results, got: %s", output)
	}
	if !strings.Contains(output, "user") {
		t.Errorf("expected type 'user' in results, got: %s", output)
	}
}

func TestSearchRecipientsJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"search", "recipients", "--search", "alice", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"name"`) {
		t.Errorf("JSON output should contain name field, got: %s", output)
	}
	if !strings.Contains(output, "Alice Student") {
		t.Errorf("JSON output should contain result name, got: %s", output)
	}
}

func TestSearchRecipientsMissingSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"search", "recipients"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --search is missing")
	}
	if !strings.Contains(execErr.Error(), "--search") {
		t.Errorf("error should mention --search, got: %v", execErr)
	}
}

func TestSearchCoursesText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"search", "courses", "--search", "CS"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "Intro to CS") {
		t.Errorf("expected 'Intro to CS' in results, got: %s", output)
	}
	if !strings.Contains(output, "CS101") {
		t.Errorf("expected course code in results, got: %s", output)
	}
}

func TestSearchAllText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(factory))

	output := captureStdout(t, func() {
		root.SetArgs([]string{"search", "all", "--search", "alice"})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// The mock /search/all returns 404, so it falls back to /search/recipients.
	if !strings.Contains(output, "Alice Student") {
		t.Errorf("expected 'Alice Student' in fallback results, got: %s", output)
	}
}

func TestSearchAllMissingSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"search", "all"})
		execErr = root.Execute()
	})

	if execErr == nil {
		t.Error("expected error when --search is missing")
	}
	if !strings.Contains(execErr.Error(), "--search") {
		t.Errorf("error should mention --search, got: %v", execErr)
	}
}
