package instagram

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSearchUsersTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "users", "--query=test")
	mustContain(t, out, "search_result")
}

func TestSearchUsersJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "--json", "search", "users", "--query=test")
	var result []UserSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one user")
	}
	if result[0].Username != "search_result" {
		t.Errorf("expected username search_result, got %s", result[0].Username)
	}
}

func TestSearchTagsTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "tags", "--query=golang")
	mustContain(t, out, "golang")
}

func TestSearchTagsJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "--json", "search", "tags", "--query=golang")
	var result []TagSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one tag")
	}
}

func TestSearchLocationsTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "locations", "--query=test")
	mustContain(t, out, "Test Location")
}

func TestSearchLocationsJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "--json", "search", "locations", "--query=test")
	var result []LocationSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one location")
	}
}

func TestSearchLocationsWithLatLng(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "locations", "--query=test", "--lat=37.7749", "--lng=-122.4194")
	mustContain(t, out, "Test Location")
}

func TestSearchTopTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "top", "--query=test")
	mustContain(t, out, "user")
}

func TestSearchTopJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "--json", "search", "top", "--query=test")
	var result []topSearchResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
}

func TestSearchClearDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "clear", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestSearchClearTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "clear")
	mustContain(t, out, "cleared")
}

func TestSearchClearJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "--json", "search", "clear")
	var result clearSearchResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status ok, got %s", result.Status)
	}
}

func TestSearchExploreTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "search", "explore")
	mustContain(t, out, "Explore items:")
}

func TestSearchExploreJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "--json", "search", "explore")
	var result exploreResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestSearchAliasFind(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSearchCmd(factory))

	out := runCmd(t, root, "find", "users", "--query=test")
	mustContain(t, out, "search_result")
}
