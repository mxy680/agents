package sheets

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValuesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "get", "--id=ss1", "--range=Sheet1!A1:B2", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result CellData
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput: %s", err, out)
	}
	if result.Range != "Sheet1!A1:B2" {
		t.Errorf("expected range Sheet1!A1:B2, got %s", result.Range)
	}
	if len(result.Values) != 2 {
		t.Errorf("expected 2 rows, got %d", len(result.Values))
	}
}

func TestValuesGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "get", "--id=ss1", "--range=Sheet1!A1:B2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Range: Sheet1!A1:B2") {
		t.Errorf("expected range header in output, got: %s", out)
	}
	if !strings.Contains(out, "Name") {
		t.Errorf("expected 'Name' in output, got: %s", out)
	}
}

func TestValuesUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "update", "--id=ss1", "--range=Sheet1!A1:B2",
		`--values=[["X","Y"],["Z","W"]]`, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result UpdateResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.UpdatedCells != 4 {
		t.Errorf("expected 4 updated cells, got %d", result.UpdatedCells)
	}
}

func TestValuesUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "update", "--id=ss1", "--range=Sheet1!A1:B2",
		`--values=[["X","Y"],["Z","W"]]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Updated") && !strings.Contains(out, "4 cells") {
		t.Errorf("expected update confirmation, got: %s", out)
	}
}

func TestValuesUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "update", "--id=ss1", "--range=Sheet1!A1:B2",
		`--values=[["X"]]`, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestValuesUpdate_MissingValues(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	_, err := runCmd(t, root, "values", "update", "--id=ss1", "--range=Sheet1!A1:B2")
	if err == nil {
		t.Fatal("expected error for missing --values")
	}
	if !strings.Contains(err.Error(), "either --values or --values-file is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValuesAppend_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "append", "--id=ss1", "--range=Sheet1!A1:B2",
		`--values=[["new","row"]]`, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result AppendResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.SpreadsheetID != "ss1" {
		t.Errorf("expected spreadsheetId ss1, got %s", result.SpreadsheetID)
	}
}

func TestValuesAppend_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "append", "--id=ss1", "--range=Sheet1!A1:B2",
		`--values=[["a"]]`, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestValuesClear_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "clear", "--id=ss1", "--range=Sheet1!A1:B2",
		"--confirm", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ClearResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result.ClearedRange != "Sheet1!A1:B2" {
		t.Errorf("expected clearedRange Sheet1!A1:B2, got %s", result.ClearedRange)
	}
}

func TestValuesClear_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	_, err := runCmd(t, root, "values", "clear", "--id=ss1", "--range=Sheet1!A1:B2")
	if err == nil {
		t.Fatal("expected error without --confirm")
	}
	if !strings.Contains(err.Error(), "irreversible") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValuesClear_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "clear", "--id=ss1", "--range=Sheet1!A1:B2", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run, got: %s", out)
	}
}

func TestValuesBatchGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	out, err := runCmd(t, root, "values", "batch-get", "--id=ss1",
		"--ranges=Sheet1!A1:B2,Sheet2!A1:B2", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []CellData
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 ranges, got %d", len(results))
	}
}

func TestValuesBatchUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	data := `[{"range":"Sheet1!A1:B2","values":[["a","b"]]},{"range":"Sheet2!A1","values":[["c"]]}]`
	out, err := runCmd(t, root, "values", "batch-update", "--id=ss1", "--data="+data, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	if result["spreadsheetId"] != "ss1" {
		t.Errorf("expected spreadsheetId ss1, got %v", result["spreadsheetId"])
	}
}

func TestValuesBatchUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	data := `[{"range":"Sheet1!A1","values":[["a"]]}]`
	out, err := runCmd(t, root, "values", "batch-update", "--id=ss1", "--data="+data, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "DRY RUN") {
		t.Errorf("expected dry run, got: %s", out)
	}
}

func TestValuesBatchUpdate_MissingData(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestValuesCmd(newTestSheetsServiceFactory(server)))

	_, err := runCmd(t, root, "values", "batch-update", "--id=ss1")
	if err == nil {
		t.Fatal("expected error for missing --data")
	}
}
