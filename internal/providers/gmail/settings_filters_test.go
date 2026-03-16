package gmail

import (
	"encoding/json"
	"strings"
	"testing"
)

// ---- settings filters list ----

func TestSettingsFiltersListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var filters []FilterInfo
	if err := json.Unmarshal([]byte(output), &filters); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(filters))
	}
	if filters[0].ID != "filter1" {
		t.Errorf("expected first filter ID=filter1, got %s", filters[0].ID)
	}
	if filters[0].Criteria.From != "newsletter@example.com" {
		t.Errorf("expected criteria.from=newsletter@example.com, got %s", filters[0].Criteria.From)
	}
}

func TestSettingsFiltersListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
	if !strings.Contains(output, "filter1") {
		t.Error("expected output to contain filter1")
	}
}

// ---- settings filters get ----

func TestSettingsFiltersGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "get", "--id=filter1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var info FilterInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if info.ID != "filter1" {
		t.Errorf("expected ID=filter1, got %s", info.ID)
	}
	if info.Criteria.From != "newsletter@example.com" {
		t.Errorf("expected criteria.from=newsletter@example.com, got %s", info.Criteria.From)
	}
}

func TestSettingsFiltersGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "get", "--id=filter1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- settings filters create ----

func TestSettingsFiltersCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "filters", "create",
			"--from=sender@example.com",
			"--add-label=STARRED",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "filter_created1" {
		t.Errorf("expected id=filter_created1, got %s", result["id"])
	}
	if result["status"] != "created" {
		t.Errorf("expected status=created, got %s", result["status"])
	}
}

func TestSettingsFiltersCreateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "filters", "create",
			"--from=sender@example.com",
			"--add-label=STARRED",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestSettingsFiltersCreateDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "filters", "create",
			"--from=sender@example.com",
			"--dry-run",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- settings filters delete ----

func TestSettingsFiltersDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "delete", "--id=filter1", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "filter1" {
		t.Errorf("expected id=filter1, got %s", result["id"])
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestSettingsFiltersDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "delete", "--id=filter1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestSettingsFiltersDeleteDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "delete", "--id=filter1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestSettingsFiltersDeleteRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	root.SetArgs([]string{"settings", "filters", "delete", "--id=filter1"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}
