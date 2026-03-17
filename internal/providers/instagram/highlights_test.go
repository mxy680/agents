package instagram

import (
	"encoding/json"
	"strings"
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

func TestHighlightsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "create", "--title=Travel", "--story-ids=story_1,story_2", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestHighlightsCreateTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "create", "--title=Travel", "--story-ids=story_1,story_2")
	mustContain(t, out, "Created highlight")
}

func TestHighlightsCreateJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "--json", "highlights", "create", "--title=Travel", "--story-ids=story_1")
	var result highlightMutateResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestHighlightsEditDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "edit", "--highlight-id=hl_111", "--title=New Title", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestHighlightsEditTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "edit", "--highlight-id=hl_111", "--title=New Title")
	mustContain(t, out, "Edited highlight")
}

func TestHighlightsEditJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "--json", "highlights", "edit", "--highlight-id=hl_111", "--add-stories=s1,s2")
	var result highlightMutateResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestHighlightsDeleteRequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	err := runCmdErr(t, root, "highlights", "delete", "--highlight-id=hl_111")
	if err == nil {
		t.Fatal("expected error without --confirm")
	}
}

func TestHighlightsDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "delete", "--highlight-id=hl_111", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestHighlightsDeleteTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestHighlightsCmd(factory))

	out := runCmd(t, root, "highlights", "delete", "--highlight-id=hl_111", "--confirm")
	mustContain(t, out, "Deleted highlight")
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
