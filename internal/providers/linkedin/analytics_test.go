package linkedin

import (
	"testing"
)

func TestAnalyticsProfileViews_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "profile-views"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "42") {
		t.Errorf("expected view count 42 in output, got: %s", out)
	}
	if !containsStr(out, "7 days") {
		t.Errorf("expected time period '7 days' in output, got: %s", out)
	}
}

func TestAnalyticsProfileViews_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "profile-views", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"total_views"`) {
		t.Errorf("expected 'total_views' field in JSON output, got: %s", out)
	}
	if !containsStr(out, "42") {
		t.Errorf("expected view count 42 in JSON output, got: %s", out)
	}
	if !containsStr(out, `"time_period"`) {
		t.Errorf("expected 'time_period' field in JSON output, got: %s", out)
	}
}

func TestAnalyticsSearchAppearances_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "search-appearances"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "15") {
		t.Errorf("expected count 15 in output, got: %s", out)
	}
	if !containsStr(out, "7 days") {
		t.Errorf("expected time period '7 days' in output, got: %s", out)
	}
}

func TestAnalyticsSearchAppearances_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "search-appearances", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"count"`) {
		t.Errorf("expected 'count' field in JSON output, got: %s", out)
	}
	if !containsStr(out, "15") {
		t.Errorf("expected count 15 in JSON output, got: %s", out)
	}
	if !containsStr(out, `"time_period"`) {
		t.Errorf("expected 'time_period' field in JSON output, got: %s", out)
	}
}

func TestAnalyticsPostImpressions_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "post-impressions", "--post-urn", "urn:li:activity:1001"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "500") {
		t.Errorf("expected impressions 500 in output, got: %s", out)
	}
	if !containsStr(out, "300") {
		t.Errorf("expected unique impressions 300 in output, got: %s", out)
	}
	if !containsStr(out, "urn:li:activity:1001") {
		t.Errorf("expected post URN in output, got: %s", out)
	}
}

func TestAnalyticsPostImpressions_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"analytics", "post-impressions", "--post-urn", "urn:li:activity:1001", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"impressions"`) {
		t.Errorf("expected 'impressions' field in JSON output, got: %s", out)
	}
	if !containsStr(out, "500") {
		t.Errorf("expected impressions 500 in JSON output, got: %s", out)
	}
	if !containsStr(out, `"unique_impressions"`) {
		t.Errorf("expected 'unique_impressions' field in JSON output, got: %s", out)
	}
	if !containsStr(out, "300") {
		t.Errorf("expected unique impressions 300 in JSON output, got: %s", out)
	}
}

func TestAnalyticsPostImpressions_MissingPostURN(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newAnalyticsCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"analytics", "post-impressions"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --post-urn is missing")
	}
}
