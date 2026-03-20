package linkedin

import (
	"encoding/json"
	"testing"
)

func TestSearchPeople_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "people", "--query", "engineer"})
		root.Execute() //nolint:errcheck
	})

	// The mock returns a company result from withCompaniesMock's /voyager/api/search/dash/clusters.
	// We verify the command runs and produces tabular output (header line).
	if !containsStr(out, "URN") && !containsStr(out, "No results found.") {
		t.Errorf("expected search output, got: %s", out)
	}
}

func TestSearchPeople_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "people", "--query", "engineer", "--json"})
		root.Execute() //nolint:errcheck
	})

	// JSON output must be an array.
	if !containsStr(out, "[") {
		t.Errorf("expected JSON array in output, got: %s", out)
	}
}

func TestSearchPeople_MissingQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"search", "people"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --query is missing")
	}
	if !containsStr(err.Error(), "--query") {
		t.Errorf("expected '--query' in error message, got: %s", err.Error())
	}
}

func TestSearchCompanies_MissingQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"search", "companies"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --query is missing")
	}
}

func TestSearchJobs_MissingQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"search", "jobs"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --query is missing")
	}
}

func TestSearchPosts_MissingQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"search", "posts"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --query is missing")
	}
}

func TestSearchGroups_MissingQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"search", "groups"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --query is missing")
	}
}

func TestSearchAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"find", "people", "--query", "test"})
		root.Execute() //nolint:errcheck
	})

	// Should produce output without an error — either table or "No results found."
	if !containsStr(out, "URN") && !containsStr(out, "No results found.") {
		t.Errorf("expected output via 'find' alias, got: %s", out)
	}
}

func TestSearchCompanies_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "companies", "--query", "software"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "URN") && !containsStr(out, "No results found.") {
		t.Errorf("expected search output, got: %s", out)
	}
}

func TestExtractGraphQLSearchResults_Empty(t *testing.T) {
	results := extractGraphQLSearchResults(nil, "person")
	if len(results) != 0 {
		t.Errorf("expected 0 results from nil included, got %d", len(results))
	}
}

func TestExtractGraphQLSearchResults_EntityResultViewModel(t *testing.T) {
	included := []json.RawMessage{
		json.RawMessage(`{
			"$type": "com.linkedin.voyager.search.dash.clusters.EntityResultViewModel",
			"entityUrn": "urn:li:fs_miniProfile:TestPerson",
			"trackingUrn": "urn:li:fs_miniProfile:TestPerson",
			"title": {"text": "John Doe"},
			"primarySubtitle": {"text": "Engineer"}
		}`),
	}
	results := extractGraphQLSearchResults(included, "person")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].URN != "urn:li:fs_miniProfile:TestPerson" {
		t.Errorf("expected URN 'urn:li:fs_miniProfile:TestPerson', got %s", results[0].URN)
	}
	if results[0].Title != "John Doe" {
		t.Errorf("expected title 'John Doe', got %s", results[0].Title)
	}
	if results[0].Type != "person" {
		t.Errorf("expected type 'person', got %s", results[0].Type)
	}
}

func TestExtractGraphQLSearchResults_FallbackToTrackingURN(t *testing.T) {
	included := []json.RawMessage{
		json.RawMessage(`{
			"$type": "com.linkedin.voyager.search.dash.clusters.EntityResultViewModel",
			"trackingUrn": "urn:li:fs_miniProfile:TrackOnly",
			"title": {"text": "Track User"},
			"primarySubtitle": {"text": "Designer"}
		}`),
	}
	results := extractGraphQLSearchResults(included, "company")
	if len(results) != 1 {
		t.Fatalf("expected 1 result via trackingUrn fallback, got %d", len(results))
	}
	if results[0].URN != "urn:li:fs_miniProfile:TrackOnly" {
		t.Errorf("expected URN 'urn:li:fs_miniProfile:TrackOnly', got %s", results[0].URN)
	}
}

func TestExtractGraphQLSearchResults_SkipsNoURN(t *testing.T) {
	included := []json.RawMessage{
		json.RawMessage(`{
			"$type": "com.linkedin.voyager.search.dash.clusters.EntityResultViewModel",
			"title": {"text": "No URN Entity"},
			"primarySubtitle": {"text": "No URN"}
		}`),
	}
	results := extractGraphQLSearchResults(included, "person")
	if len(results) != 0 {
		t.Errorf("expected 0 results when entityUrn and trackingUrn are empty, got %d", len(results))
	}
}

func TestSearchPeople_InvalidCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newSearchCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"search", "people", "--query", "test", "--cursor", "notanumber"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid cursor")
	}
}
