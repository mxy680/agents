package linkedin

import (
	"context"
	"testing"
)

func TestEventsList_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"events", "list"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Go Meetup 2024") {
		t.Errorf("expected 'Go Meetup 2024' in output, got: %s", out)
	}
	if !containsStr(out, "Cloud Summit") {
		t.Errorf("expected 'Cloud Summit' in output, got: %s", out)
	}
}

func TestEventsList_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "events", "list"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"title"`) {
		t.Errorf("expected JSON field 'title' in output, got: %s", out)
	}
	if !containsStr(out, "Go Meetup 2024") {
		t.Errorf("expected event title in JSON output, got: %s", out)
	}
	if !containsStr(out, `"starts_at"`) {
		t.Errorf("expected 'starts_at' in JSON output, got: %s", out)
	}
}

func TestEventsList_WithAlias(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"event", "list"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Go Meetup 2024") {
		t.Errorf("expected event title in output via alias, got: %s", out)
	}
}

func TestEventsGet_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"events", "get", "--event-id=67890"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Go Meetup 2024") {
		t.Errorf("expected event title in output, got: %s", out)
	}
	if !containsStr(out, "San Francisco") {
		t.Errorf("expected location in output, got: %s", out)
	}
}

func TestEventsGet_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "events", "get", "--event-id=67890"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"title"`) {
		t.Errorf("expected JSON field 'title' in output, got: %s", out)
	}
	if !containsStr(out, "Go Meetup 2024") {
		t.Errorf("expected event title in JSON output, got: %s", out)
	}
}

func TestEventsGet_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"events", "get"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --event-id is missing")
	}
}

func TestEventsAttend_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"events", "attend", "--event-id=67890", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestEventsAttend_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"events", "attend", "--event-id=67890"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "67890") {
		t.Errorf("expected event ID in output, got: %s", out)
	}
}

func TestEventsAttend_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "events", "attend", "--event-id=67890"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"attending"`) {
		t.Errorf("expected JSON field 'attending' in output, got: %s", out)
	}
}

func TestEventsAttend_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"events", "attend"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --event-id is missing")
	}
}

func TestEventsUnattend_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"events", "unattend", "--event-id=67890", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestEventsUnattend_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"events", "unattend", "--event-id=67890"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "67890") {
		t.Errorf("expected event ID in output, got: %s", out)
	}
}

func TestEventsUnattend_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"events", "unattend"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --event-id is missing")
	}
}

func TestToEventSummary(t *testing.T) {
	el := voyagerEventElement{
		EntityURN: "urn:li:fs_event:67890",
		Title:     "Go Meetup 2024",
		StartAt:   1704067200000,
		Location:  "San Francisco, CA",
	}
	s := toEventSummary(el)
	if s.ID != "67890" {
		t.Errorf("ID = %q, want %q", s.ID, "67890")
	}
	if s.Title != "Go Meetup 2024" {
		t.Errorf("Title = %q, want %q", s.Title, "Go Meetup 2024")
	}
	if s.StartsAt != 1704067200000 {
		t.Errorf("StartsAt = %d, want 1704067200000", s.StartsAt)
	}
	if s.Location != "San Francisco, CA" {
		t.Errorf("Location = %q, want %q", s.Location, "San Francisco, CA")
	}
}
