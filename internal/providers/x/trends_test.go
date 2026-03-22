package x

import (
	"testing"
)

func TestTrendsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newTrendsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestTrendsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newTrendsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "list"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestTrendsLocations_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newTrendsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "locations", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestTrendsLocations_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newTrendsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "locations"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestTrendsByPlace_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newTrendsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "by-place", "--woeid=1", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestTrendsByPlace_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newTrendsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trends", "by-place", "--woeid=1"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestTrendsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newTrendsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"trend", "list", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output via 'trend' alias, got empty string")
	}
}

func TestExtractTrendsFromGuide(t *testing.T) {
	// Test with empty data — should not panic.
	result := extractTrendsFromGuide([]byte(`{}`))
	if result == nil {
		// nil is acceptable for empty data.
		return
	}
	if len(result) != 0 {
		t.Errorf("expected no trends from empty guide, got %d", len(result))
	}
}

func TestPrintTrendSummaries_Empty(t *testing.T) {
	root := newTestRootCmd()
	// Provide a subcommand that calls printTrendSummaries with an empty list.
	called := false
	out := captureStdout(t, func() {
		called = true
		err := printTrendSummaries(root, []TrendSummary{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !called {
		t.Error("expected printTrendSummaries to be called")
	}
	if !containsStr(out, "No trends") {
		t.Errorf("expected 'No trends' in output, got: %s", out)
	}
}
