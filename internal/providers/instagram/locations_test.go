package instagram

import (
	"encoding/json"
	"testing"
)

func TestLocationsGetTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "locations", "get", "--location-id=999111")
	mustContain(t, out, "Test Location")
	mustContain(t, out, "Posts:")
}

func TestLocationsGetJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "--json", "locations", "get", "--location-id=999111")
	var result LocationSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
	if result.Name == "" {
		t.Error("expected non-empty location name")
	}
}

func TestLocationsFeedTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "locations", "feed", "--location-id=999111")
	mustContain(t, out, "loc_media_111")
}

func TestLocationsFeedJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "--json", "locations", "feed", "--location-id=999111")
	var result []MediaSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one media item")
	}
}

func TestLocationsFeedWithTab(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "locations", "feed", "--location-id=999111", "--tab=recent")
	mustContain(t, out, "loc_media_111")
}

func TestLocationsSearchTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "locations", "search", "--query=test")
	mustContain(t, out, "Test Location")
}

func TestLocationsSearchJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "--json", "locations", "search", "--query=test")
	var result []LocationSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one location")
	}
}

func TestLocationsSearchWithCoords(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "locations", "search", "--query=test", "--lat=37.7749", "--lng=-122.4194")
	mustContain(t, out, "Test Location")
}

func TestLocationsStoriesTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "locations", "stories", "--location-id=999111")
	mustContain(t, out, "retrieved")
}

func TestLocationsStoriesJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLocationsCmd(factory))

	out := runCmd(t, root, "--json", "locations", "stories", "--location-id=999111")
	var result locationStoryResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestLocationsAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)

	for _, alias := range []string{"location", "loc"} {
		root := newTestRootCmd()
		root.AddCommand(buildTestLocationsCmd(factory))
		out := runCmd(t, root, alias, "get", "--location-id=999111")
		mustContain(t, out, "Test Location")
	}
}
