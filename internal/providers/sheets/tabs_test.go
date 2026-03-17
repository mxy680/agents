package sheets

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTabsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "tabs", "list", "--id=ss1", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var tabs []SheetInfo
	if err := json.Unmarshal([]byte(out), &tabs); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if len(tabs) != 2 {
		t.Errorf("expected 2 tabs, got %d", len(tabs))
	}
	if tabs[0].Title != "Sheet1" {
		t.Errorf("expected first tab 'Sheet1', got %s", tabs[0].Title)
	}
}

func TestTabsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "tabs", "list", "--id=ss1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Sheet1") {
		t.Errorf("expected Sheet1 in output, got: %s", out)
	}
	if !strings.Contains(out, "Sheet2") {
		t.Errorf("expected Sheet2 in output, got: %s", out)
	}
}

func TestTabsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "tabs", "create", "--id=ss1", "--title=NewTab", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SheetInfo
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.SheetID != 42 {
		t.Errorf("expected sheetId 42, got %d", result.SheetID)
	}
}

func TestTabsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "tabs", "create", "--id=ss1", "--title=NewTab", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run, got: %s", out)
	}
}

func TestTabsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "tabs", "delete", "--id=ss1", "--sheet-id=0", "--confirm", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status deleted, got %v", result["status"])
	}
}

func TestTabsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	_, err := runCmd(t, root, "tabs", "delete", "--id=ss1", "--sheet-id=0")
	if err == nil {
		t.Fatal("expected error without --confirm")
	}
	if !strings.Contains(err.Error(), "irreversible") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTabsRename_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "tabs", "rename", "--id=ss1", "--sheet-id=0", "--title=Renamed", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result["status"] != "renamed" {
		t.Errorf("expected status renamed, got %v", result["status"])
	}
}

func TestTabsRename_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestTabsCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "tabs", "rename", "--id=ss1", "--sheet-id=0", "--title=X", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run, got: %s", out)
	}
}
