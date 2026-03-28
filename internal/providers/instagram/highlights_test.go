package instagram

import (
	"encoding/json"
	"testing"
)

func TestHighlightsListTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "list")
	mustContain(t, out, "Travel")
}

func TestHighlightsListJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "--json", "highlights", "list")
	var result []HighlightSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one highlight")
	}
	if result[0].Title != "Travel" {
		t.Errorf("expected title Travel, got %s", result[0].Title)
	}
}

func TestHighlightsListWithUserID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "list", "--user-id=42544748138")
	mustContain(t, out, "Travel")
}

func TestHighlightsGetTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "get", "--highlight-id=hl_111")
	mustContain(t, out, "retrieved")
}

func TestHighlightsGetJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "--json", "highlights", "get", "--highlight-id=hl_111")
	var result highlightMediaResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}


func TestHighlightsAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)

	for _, alias := range []string{"highlight", "hl"} {
		root := newTestRootCmd()
		root.AddCommand(buildTestHighlightsCmd(factory))
		out := runCmd(t, root, alias, "list")
		mustContain(t, out, "Travel")
	}
}
