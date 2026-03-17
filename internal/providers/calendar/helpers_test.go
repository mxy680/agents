package calendar

import (
	"testing"

	"github.com/spf13/cobra"
	api "google.golang.org/api/calendar/v3"
)

func TestFormatEventTime(t *testing.T) {
	tests := []struct {
		name string
		edt  *api.EventDateTime
		want string
	}{
		{"nil", nil, ""},
		{"dateTime", &api.EventDateTime{DateTime: "2026-03-16T09:00:00-04:00"}, "2026-03-16T09:00:00-04:00"},
		{"date", &api.EventDateTime{Date: "2026-03-16"}, "2026-03-16"},
		{"both prefers dateTime", &api.EventDateTime{DateTime: "2026-03-16T09:00:00Z", Date: "2026-03-16"}, "2026-03-16T09:00:00Z"},
		{"empty", &api.EventDateTime{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEventTime(tt.edt)
			if got != tt.want {
				t.Errorf("formatEventTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToEventSummary(t *testing.T) {
	event := &api.Event{
		Id:       "ev1",
		Summary:  "Meeting",
		Location: "Room A",
		Status:   "confirmed",
		Start:    &api.EventDateTime{DateTime: "2026-03-16T09:00:00Z"},
		End:      &api.EventDateTime{DateTime: "2026-03-16T10:00:00Z"},
		Organizer: &api.EventOrganizer{Email: "org@example.com"},
	}

	s := toEventSummary(event)
	if s.ID != "ev1" {
		t.Errorf("expected ID=ev1, got %s", s.ID)
	}
	if s.Summary != "Meeting" {
		t.Errorf("expected Summary=Meeting, got %s", s.Summary)
	}
	if s.Organizer != "org@example.com" {
		t.Errorf("expected Organizer=org@example.com, got %s", s.Organizer)
	}
	if s.Start != "2026-03-16T09:00:00Z" {
		t.Errorf("expected Start=2026-03-16T09:00:00Z, got %s", s.Start)
	}
}

func TestToEventSummaryNoOrganizer(t *testing.T) {
	event := &api.Event{
		Id:     "ev2",
		Status: "confirmed",
		Start:  &api.EventDateTime{Date: "2026-03-16"},
		End:    &api.EventDateTime{Date: "2026-03-17"},
	}

	s := toEventSummary(event)
	if s.Organizer != "" {
		t.Errorf("expected empty Organizer, got %s", s.Organizer)
	}
	if s.Start != "2026-03-16" {
		t.Errorf("expected Start=2026-03-16, got %s", s.Start)
	}
}

func TestToEventDetail(t *testing.T) {
	event := &api.Event{
		Id:          "ev1",
		Summary:     "Meeting",
		Description: "A team meeting",
		Location:    "Room A",
		Status:      "confirmed",
		Start:       &api.EventDateTime{DateTime: "2026-03-16T09:00:00Z"},
		End:         &api.EventDateTime{DateTime: "2026-03-16T10:00:00Z"},
		Organizer:   &api.EventOrganizer{Email: "org@example.com"},
		HangoutLink: "https://meet.google.com/abc",
		HtmlLink:    "https://calendar.google.com/event?eid=ev1",
		Created:     "2026-03-01T10:00:00Z",
		Updated:     "2026-03-15T10:00:00Z",
		Attendees: []*api.EventAttendee{
			{Email: "bob@example.com", ResponseStatus: "accepted", Self: true},
			{Email: "charlie@example.com", ResponseStatus: "tentative", Organizer: true},
		},
	}

	d := toEventDetail(event)
	if d.ID != "ev1" {
		t.Errorf("expected ID=ev1, got %s", d.ID)
	}
	if d.Description != "A team meeting" {
		t.Errorf("expected Description, got %s", d.Description)
	}
	if d.HangoutLink != "https://meet.google.com/abc" {
		t.Errorf("expected HangoutLink, got %s", d.HangoutLink)
	}
	if len(d.Attendees) != 2 {
		t.Fatalf("expected 2 attendees, got %d", len(d.Attendees))
	}
	if d.Attendees[0].Email != "bob@example.com" {
		t.Errorf("expected first attendee bob, got %s", d.Attendees[0].Email)
	}
	if !d.Attendees[0].Self {
		t.Error("expected first attendee Self=true")
	}
	if !d.Attendees[1].Organizer {
		t.Error("expected second attendee Organizer=true")
	}
}

func TestToEventDetailNoOrganizer(t *testing.T) {
	event := &api.Event{
		Id:    "ev2",
		Start: &api.EventDateTime{},
		End:   &api.EventDateTime{},
	}
	d := toEventDetail(event)
	if d.Organizer != "" {
		t.Errorf("expected empty Organizer, got %s", d.Organizer)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}

func TestParseAttendees(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"single", "alice@example.com", 1},
		{"multiple", "alice@example.com, bob@example.com, charlie@example.com", 3},
		{"with spaces", " alice@example.com , bob@example.com ", 2},
		{"trailing comma", "alice@example.com,", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAttendees(tt.input)
			if len(got) != tt.want {
				t.Errorf("parseAttendees(%q) returned %d attendees, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}

func TestParseAttendeesEmails(t *testing.T) {
	attendees := parseAttendees("alice@example.com, bob@example.com")
	if attendees[0].Email != "alice@example.com" {
		t.Errorf("expected first email=alice@example.com, got %s", attendees[0].Email)
	}
	if attendees[1].Email != "bob@example.com" {
		t.Errorf("expected second email=bob@example.com, got %s", attendees[1].Email)
	}
}

func TestConfirmDestructive(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("confirm", false, "")

	err := confirmDestructive(cmd)
	if err == nil {
		t.Error("expected error when --confirm not set")
	}

	cmd.Flags().Set("confirm", "true")
	err = confirmDestructive(cmd)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
