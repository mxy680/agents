package linkedin

import (
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

func TestExtractSearchResults_EmptyClusters(t *testing.T) {
	raw := voyagerSearchResponse{}
	results := extractSearchResults(raw, "person")
	if len(results) != 0 {
		t.Errorf("expected 0 results from empty response, got %d", len(results))
	}
}

func TestExtractSearchResults_Items(t *testing.T) {
	raw := voyagerSearchResponse{
		Elements: []struct {
			Items []struct {
				Item struct {
					EntityResult struct {
						EntityURN       string `json:"entityUrn"`
						Title           struct{ Text string `json:"text"` } `json:"title"`
						PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
					} `json:"entityResult"`
					SearchEntityResult struct {
						EntityURN       string `json:"entityUrn"`
						Title           struct{ Text string `json:"text"` } `json:"title"`
						PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
					} `json:"com.linkedin.voyager.search.SearchEntityResult"`
				} `json:"item"`
			} `json:"items"`
			Elements []struct {
				Item struct {
					EntityResult struct {
						EntityURN       string `json:"entityUrn"`
						Title           struct{ Text string `json:"text"` } `json:"title"`
						PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
					} `json:"entityResult"`
					SearchEntityResult struct {
						EntityURN       string `json:"entityUrn"`
						Title           struct{ Text string `json:"text"` } `json:"title"`
						PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
					} `json:"com.linkedin.voyager.search.SearchEntityResult"`
				} `json:"item"`
			} `json:"elements"`
		}{
			{
				Items: []struct {
					Item struct {
						EntityResult struct {
							EntityURN       string `json:"entityUrn"`
							Title           struct{ Text string `json:"text"` } `json:"title"`
							PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
						} `json:"entityResult"`
						SearchEntityResult struct {
							EntityURN       string `json:"entityUrn"`
							Title           struct{ Text string `json:"text"` } `json:"title"`
							PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
						} `json:"com.linkedin.voyager.search.SearchEntityResult"`
					} `json:"item"`
				}{
					{
						Item: struct {
							EntityResult struct {
								EntityURN       string `json:"entityUrn"`
								Title           struct{ Text string `json:"text"` } `json:"title"`
								PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
							} `json:"entityResult"`
							SearchEntityResult struct {
								EntityURN       string `json:"entityUrn"`
								Title           struct{ Text string `json:"text"` } `json:"title"`
								PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
							} `json:"com.linkedin.voyager.search.SearchEntityResult"`
						}{
							EntityResult: struct {
								EntityURN       string `json:"entityUrn"`
								Title           struct{ Text string `json:"text"` } `json:"title"`
								PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
							}{
								EntityURN: "urn:li:fs_miniProfile:TestPerson",
								Title:     struct{ Text string `json:"text"` }{Text: "John Doe"},
							},
						},
					},
				},
			},
		},
	}
	results := extractSearchResults(raw, "person")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].URN != "urn:li:fs_miniProfile:TestPerson" {
		t.Errorf("expected URN 'urn:li:fs_miniProfile:TestPerson', got %s", results[0].URN)
	}
	if results[0].Type != "person" {
		t.Errorf("expected type 'person', got %s", results[0].Type)
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
