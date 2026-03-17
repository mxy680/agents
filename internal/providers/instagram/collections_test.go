package instagram

import (
	"encoding/json"
	"strings"
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

func TestCollectionsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "create", "--name=Test", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestCollectionsCreateTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "create", "--name=New Collection")
	mustContain(t, out, "Created collection")
}

func TestCollectionsCreateJSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "--json", "collections", "create", "--name=New Collection")
	var result collectionMutateResponse
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerr: %v", out, err)
	}
}

func TestCollectionsEditDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "edit", "--collection-id=col_111", "--name=Renamed", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestCollectionsEditTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "edit", "--collection-id=col_111", "--name=Renamed")
	mustContain(t, out, "Renamed collection")
}

func TestCollectionsDeleteRequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	err := runCmdErr(t, root, "collections", "delete", "--collection-id=col_111")
	if err == nil {
		t.Fatal("expected error without --confirm")
	}
}

func TestCollectionsDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "delete", "--collection-id=col_111", "--dry-run")
	if !strings.Contains(out, "[DRY RUN]") {
		t.Errorf("expected dry-run output, got: %s", out)
	}
}

func TestCollectionsDeleteTextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCollectionsCmd(factory))

	out := runCmd(t, root, "collections", "delete", "--collection-id=col_111", "--confirm")
	mustContain(t, out, "Deleted collection")
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
