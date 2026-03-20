package linkedin

import (
	"context"
	"testing"
)

func TestEventsList_Deprecated(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"events", "list"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for deprecated events list endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestEventsList_AliasDeprecated(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"event", "list"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for deprecated events list via alias")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message via alias, got: %s", err.Error())
	}
}

func TestEventsGet_Deprecated(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newEventsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"events", "get", "--event-id=67890"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for deprecated events get endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
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
