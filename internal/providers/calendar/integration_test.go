//go:build integration

package calendar

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// requireEnv skips the test if any required env var is missing.
func requireEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"GMAIL_ACCESS_TOKEN",
		"GMAIL_REFRESH_TOKEN",
	} {
		if os.Getenv(key) == "" {
			t.Skipf("skipping: %s not set (run with doppler run)", key)
		}
	}
}

func realFactory() ServiceFactory {
	return auth.NewCalendarService
}

func integrationRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// --- calendars list ---

func TestIntegration_CalendarsList_JSON(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestCalendarsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("calendars list failed: %v", execErr)
	}

	var summaries []CalendarSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d calendars", len(summaries))
	for _, c := range summaries {
		t.Logf("  [%s] summary=%q primary=%v access=%s", c.ID, c.Summary, c.Primary, c.AccessRole)
	}
	if len(summaries) == 0 {
		t.Error("expected at least one calendar (primary)")
	}
}

func TestIntegration_CalendarsList_Text(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestCalendarsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("calendars list text failed: %v", execErr)
	}
	t.Logf("text output:\n%s", output)
}

// --- calendars get ---

func TestIntegration_CalendarsGet_JSON(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestCalendarsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "get", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("calendars get failed: %v", execErr)
	}

	var summary CalendarSummary
	if err := json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("primary calendar: id=%s summary=%q tz=%s", summary.ID, summary.Summary, summary.TimeZone)
}

// --- events list ---

func TestIntegration_EventsList_JSON(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestEventsCmd(realFactory()))

	now := time.Now()
	timeMin := now.Format(time.RFC3339)
	timeMax := now.Add(7 * 24 * time.Hour).Format(time.RFC3339)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"events", "list",
			"--time-min=" + timeMin,
			"--time-max=" + timeMax,
			"--single-events",
			"--order-by=startTime",
			"--limit=10",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("events list failed: %v", execErr)
	}

	var summaries []EventSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d events in next 7 days", len(summaries))
	for _, e := range summaries {
		t.Logf("  [%s] %q start=%s", e.ID, e.Summary, e.Start)
	}
}

func TestIntegration_EventsList_Text(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestEventsCmd(realFactory()))

	now := time.Now()
	timeMin := now.Format(time.RFC3339)
	timeMax := now.Add(7 * 24 * time.Hour).Format(time.RFC3339)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"events", "list",
			"--time-min=" + timeMin,
			"--time-max=" + timeMax,
			"--single-events",
			"--order-by=startTime",
			"--limit=10",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("events list text failed: %v", execErr)
	}
	t.Logf("text output:\n%s", output)
}

// --- events get ---

func TestIntegration_EventsGet_JSON(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	now := time.Now()
	resp, err := svc.Events.List("primary").
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(now.Add(7 * 24 * time.Hour).Format(time.RFC3339)).
		SingleEvents(true).
		MaxResults(1).
		Do()
	if err != nil {
		t.Fatalf("listing events: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Skip("no events in next 7 days")
	}
	eventID := resp.Items[0].Id

	root := integrationRootCmd()
	root.AddCommand(buildTestEventsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"events", "get", "--event-id=" + eventID, "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("events get failed: %v", execErr)
	}

	var detail EventDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if detail.ID != eventID {
		t.Errorf("expected ID=%s, got %s", eventID, detail.ID)
	}
	t.Logf("event: summary=%q start=%s attendees=%d", detail.Summary, detail.Start, len(detail.Attendees))
}

// --- events create (dry-run) ---

func TestIntegration_EventsCreate_DryRun(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestEventsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"events", "create",
			"--summary=Integration Test Event (dry-run)",
			"--start=2026-03-20T10:00:00Z",
			"--end=2026-03-20T11:00:00Z",
			"--description=This should NOT be created.",
			"--dry-run",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("events create dry-run failed: %v", execErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["action"] != "create" {
		t.Errorf("expected action=create, got %v", result["action"])
	}
	t.Logf("dry-run result: %v", result)
}

// --- events create + get + update + delete (full lifecycle) ---

func TestIntegration_EventLifecycle(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	// Create
	root := integrationRootCmd()
	root.AddCommand(buildTestEventsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"events", "create",
			"--summary=CI Lifecycle Test " + time.Now().Format("15:04:05"),
			"--start=" + time.Now().Add(48*time.Hour).Format(time.RFC3339),
			"--end=" + time.Now().Add(49*time.Hour).Format(time.RFC3339),
			"--description=Auto-created by integration test. Safe to delete.",
			"--json",
		})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("events create failed: %v", execErr)
	}

	var created EventDetail
	if err := json.Unmarshal([]byte(output), &created); err != nil {
		t.Fatalf("invalid JSON from create: %v\noutput: %s", err, output)
	}
	eventID := created.ID
	t.Logf("created event: id=%s summary=%q", eventID, created.Summary)

	// Cleanup: always delete the event
	defer func() {
		if err := svc.Events.Delete("primary", eventID).Do(); err != nil {
			t.Logf("WARNING: failed to clean up event %s: %v", eventID, err)
		} else {
			t.Logf("cleaned up event %s", eventID)
		}
	}()

	// Get
	root2 := integrationRootCmd()
	root2.AddCommand(buildTestEventsCmd(realFactory()))
	output = captureStdout(t, func() {
		root2.SetArgs([]string{"events", "get", "--event-id=" + eventID, "--json"})
		execErr = root2.Execute()
	})
	if execErr != nil {
		t.Fatalf("events get failed: %v", execErr)
	}

	var fetched EventDetail
	if err := json.Unmarshal([]byte(output), &fetched); err != nil {
		t.Fatalf("invalid JSON from get: %v", err)
	}
	if fetched.ID != eventID {
		t.Errorf("get: expected ID=%s, got %s", eventID, fetched.ID)
	}
	t.Logf("fetched event: summary=%q description=%q", fetched.Summary, fetched.Description)

	// Update
	root3 := integrationRootCmd()
	root3.AddCommand(buildTestEventsCmd(realFactory()))
	output = captureStdout(t, func() {
		root3.SetArgs([]string{"events", "update",
			"--event-id=" + eventID,
			"--summary=Updated Lifecycle Test",
			"--location=Test Room 42",
			"--json",
		})
		execErr = root3.Execute()
	})
	if execErr != nil {
		t.Fatalf("events update failed: %v", execErr)
	}

	var updated EventDetail
	if err := json.Unmarshal([]byte(output), &updated); err != nil {
		t.Fatalf("invalid JSON from update: %v", err)
	}
	t.Logf("updated event: summary=%q location=%q", updated.Summary, updated.Location)

	// Delete via CLI (confirm)
	root4 := integrationRootCmd()
	root4.AddCommand(buildTestEventsCmd(realFactory()))
	output = captureStdout(t, func() {
		root4.SetArgs([]string{"events", "delete", "--event-id=" + eventID, "--confirm", "--json"})
		execErr = root4.Execute()
	})
	if execErr != nil {
		t.Fatalf("events delete failed: %v", execErr)
	}

	var deleted map[string]string
	if err := json.Unmarshal([]byte(output), &deleted); err != nil {
		t.Fatalf("invalid JSON from delete: %v", err)
	}
	if deleted["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", deleted["status"])
	}
	t.Logf("deleted event: %s", eventID)
}

// --- events quick-add (dry-run) ---

func TestIntegration_EventsQuickAdd_DryRun(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestEventsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"events", "quick-add",
			"--text=Integration test meeting tomorrow at 3pm",
			"--dry-run",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("events quick-add dry-run failed: %v", execErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["action"] != "quick-add" {
		t.Errorf("expected action=quick-add, got %v", result["action"])
	}
	t.Logf("dry-run result: %v", result)
}

// --- freebusy query ---

func TestIntegration_FreebusyQuery_JSON(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(buildTestFreebusyCmd(realFactory()))

	now := time.Now()
	timeMin := now.Format(time.RFC3339)
	timeMax := now.Add(24 * time.Hour).Format(time.RFC3339)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"freebusy", "query",
			"--time-min=" + timeMin,
			"--time-max=" + timeMax,
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("freebusy query failed: %v", execErr)
	}

	var results []FreeBusyResult
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d calendar results", len(results))
	for _, r := range results {
		t.Logf("  calendar=%s busy_slots=%d", r.CalendarID, len(r.Busy))
		for _, b := range r.Busy {
			t.Logf("    %s → %s", b.Start, b.End)
		}
	}
}

// --- calendars create + update + delete (lifecycle, uses secondary calendar) ---

func TestIntegration_CalendarLifecycle(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	// Create
	root := integrationRootCmd()
	root.AddCommand(buildTestCalendarsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"calendars", "create",
			"--summary=CI Test Calendar " + time.Now().Format("15:04:05"),
			"--description=Auto-created by integration test. Safe to delete.",
			"--timezone=America/New_York",
			"--json",
		})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("calendars create failed: %v", execErr)
	}

	var created CalendarSummary
	if err := json.Unmarshal([]byte(output), &created); err != nil {
		t.Fatalf("invalid JSON from create: %v\noutput: %s", err, output)
	}
	calID := created.ID
	t.Logf("created calendar: id=%s summary=%q", calID, created.Summary)

	// Cleanup: always delete
	defer func() {
		if err := svc.Calendars.Delete(calID).Do(); err != nil {
			t.Logf("WARNING: failed to clean up calendar %s: %v", calID, err)
		} else {
			t.Logf("cleaned up calendar %s", calID)
		}
	}()

	// Update
	root2 := integrationRootCmd()
	root2.AddCommand(buildTestCalendarsCmd(realFactory()))
	output = captureStdout(t, func() {
		root2.SetArgs([]string{"calendars", "update",
			"--calendar-id=" + calID,
			"--summary=Updated CI Calendar",
			"--json",
		})
		execErr = root2.Execute()
	})
	if execErr != nil {
		t.Fatalf("calendars update failed: %v", execErr)
	}
	t.Logf("updated calendar %s", calID)

	// Delete via CLI
	root3 := integrationRootCmd()
	root3.AddCommand(buildTestCalendarsCmd(realFactory()))
	output = captureStdout(t, func() {
		root3.SetArgs([]string{"calendars", "delete",
			"--calendar-id=" + calID,
			"--confirm",
			"--json",
		})
		execErr = root3.Execute()
	})
	if execErr != nil {
		t.Fatalf("calendars delete failed: %v", execErr)
	}

	var deleted map[string]string
	if err := json.Unmarshal([]byte(output), &deleted); err != nil {
		t.Fatalf("invalid JSON from delete: %v", err)
	}
	if deleted["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", deleted["status"])
	}
	t.Logf("deleted calendar: %s", calID)
}
