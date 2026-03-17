package calendar

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---- calendars list ----

func TestCalendarsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var summaries []CalendarSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 calendars, got %d", len(summaries))
	}
	if summaries[0].ID != "primary" {
		t.Errorf("expected first calendar ID=primary, got %s", summaries[0].ID)
	}
	if !summaries[0].Primary {
		t.Error("expected first calendar Primary=true")
	}
}

func TestCalendarsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestCalendarsListWithFlags(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "list", "--limit=5", "--show-hidden"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}

// ---- calendars get ----

func TestCalendarsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "get", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var summary CalendarSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if summary.ID != "primary" {
		t.Errorf("expected ID=primary, got %s", summary.ID)
	}
	if summary.Summary != "My Calendar" {
		t.Errorf("expected Summary=My Calendar, got %s", summary.Summary)
	}
}

func TestCalendarsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "get"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- calendars create ----

func TestCalendarsCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "create",
			"--summary=New Calendar",
			"--timezone=America/New_York",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var summary CalendarSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if summary.ID != "new-cal-123" {
		t.Errorf("expected ID=new-cal-123, got %s", summary.ID)
	}
}

func TestCalendarsCreateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "create", "--summary=New Calendar"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestCalendarsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "create", "--summary=Dry Run Calendar", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- calendars update ----

func TestCalendarsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "update",
			"--summary=Updated Calendar",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var summary CalendarSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
}

func TestCalendarsUpdateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "update", "--summary=Dry Run", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- calendars delete ----

func TestCalendarsDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "delete", "--calendar-id=work-cal-1", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestCalendarsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	root.SetArgs([]string{"calendars", "delete", "--calendar-id=work-cal-1"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not set")
	}
}

func TestCalendarsDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "delete", "--calendar-id=work-cal-1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestCalendarsUpdateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "update",
			"--summary=Updated Name",
			"--description=Updated desc",
			"--timezone=America/Chicago",
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

func TestCalendarsListEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/me/calendarList", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"items": []map[string]any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected output for empty list")
	}
}

func TestCalendarsDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCalendarsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "delete", "--calendar-id=work-cal-1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}
