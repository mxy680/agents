package sheets

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSpreadsheetsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "list", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}

	spreadsheets, ok := result["spreadsheets"].([]any)
	if !ok {
		t.Fatal("expected spreadsheets array in output")
	}
	if len(spreadsheets) != 2 {
		t.Errorf("expected 2 spreadsheets, got %d", len(spreadsheets))
	}
}

func TestSpreadsheetsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Test Spreadsheet") {
		t.Errorf("expected 'Test Spreadsheet' in output, got: %s", out)
	}
	if !strings.Contains(out, "Another Sheet") {
		t.Errorf("expected 'Another Sheet' in output, got: %s", out)
	}
}

func TestSpreadsheetsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "get", "--id=ss1", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SpreadsheetDetail
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.Title != "Test Spreadsheet" {
		t.Errorf("expected title 'Test Spreadsheet', got %s", result.Title)
	}
	if len(result.Sheets) != 2 {
		t.Errorf("expected 2 sheets, got %d", len(result.Sheets))
	}
}

func TestSpreadsheetsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "get", "--id=ss1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Test Spreadsheet") {
		t.Errorf("expected title in output, got: %s", out)
	}
	if !strings.Contains(out, "Sheet1") {
		t.Errorf("expected Sheet1 in output, got: %s", out)
	}
}

func TestSpreadsheetsCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "create", "--title=My New Sheet", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result SpreadsheetSummary
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.ID != "new-ss-id" {
		t.Errorf("expected id 'new-ss-id', got %s", result.ID)
	}
}

func TestSpreadsheetsCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "create", "--title=Test", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run, got: %s", out)
	}
}

func TestSpreadsheetsDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "delete", "--id=ss1", "--confirm", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status deleted, got %s", result["status"])
	}
}

func TestSpreadsheetsDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	_, err := runCmd(t, root, "spreadsheets", "delete", "--id=ss1")
	if err == nil {
		t.Fatal("expected error without --confirm")
	}
	if !strings.Contains(err.Error(), "irreversible") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSpreadsheetsDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestSpreadsheetsCmd(
		newTestSheetsServiceFactory(server),
		newTestDriveServiceFactory(server),
	))

	out, err := runCmd(t, root, "spreadsheets", "delete", "--id=ss1", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run, got: %s", out)
	}
}
