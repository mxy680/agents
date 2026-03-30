package instagram

import (
	"encoding/json"
	"testing"
)

func TestCollectionsListTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "list")
	mustContain(t, out, "My Saves")
}

func TestCollectionsListJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "--json", "collections", "list")
	var result []CollectionSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one collection")
	}
	if result[0].CollectionID != "col_111" {
		t.Errorf("expected collection_id col_111, got %s", result[0].CollectionID)
	}
}

func TestCollectionsGetTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "get", "--collection-id=col_111")
	mustContain(t, out, "saved_media_111")
}

func TestCollectionsGetJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "--json", "collections", "get", "--collection-id=col_111")
	var result []MediaSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one media item")
	}
}


func TestCollectionsSavedTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "saved")
	mustContain(t, out, "saved_post_222")
}

func TestCollectionsSavedJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "--json", "collections", "saved")
	var result []MediaSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerr: %v", out, err)
	}
}

func TestCollectionsAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)

	for _, alias := range []string{"collection", "saved"} {
		root := newTestRootCmd()
		root.AddCommand(buildTestCollectionsCmd(factory))
		out := runCmd(t, root, alias, "list")
		mustContain(t, out, "My Saves")
	}
}
