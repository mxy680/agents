package calendar

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/calendar/v3"
)

// EventSummary is the JSON-serializable summary of a calendar event.
type EventSummary struct {
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Location  string `json:"location,omitempty"`
	Status    string `json:"status"`
	Organizer string `json:"organizer,omitempty"`
}

// EventDetail is the JSON-serializable full event content.
type EventDetail struct {
	ID          string         `json:"id"`
	Summary     string         `json:"summary"`
	Description string         `json:"description,omitempty"`
	Start       string         `json:"start"`
	End         string         `json:"end"`
	Location    string         `json:"location,omitempty"`
	Status      string         `json:"status"`
	Organizer   string         `json:"organizer,omitempty"`
	Attendees   []AttendeeInfo `json:"attendees,omitempty"`
	HangoutLink string         `json:"hangoutLink,omitempty"`
	HTMLURL     string         `json:"htmlLink,omitempty"`
	Created     string         `json:"created,omitempty"`
	Updated     string         `json:"updated,omitempty"`
}

// AttendeeInfo is a JSON-serializable attendee entry.
type AttendeeInfo struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
	Organizer      bool   `json:"organizer,omitempty"`
	Self           bool   `json:"self,omitempty"`
}

// CalendarSummary is the JSON-serializable summary of a calendar.
type CalendarSummary struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	TimeZone    string `json:"timeZone,omitempty"`
	Primary     bool   `json:"primary,omitempty"`
	AccessRole  string `json:"accessRole,omitempty"`
}

// FreeBusyResult is the JSON-serializable free/busy result for one calendar.
type FreeBusyResult struct {
	CalendarID string     `json:"calendarId"`
	Busy       []BusySlot `json:"busy"`
}

// BusySlot is a time range during which the calendar is busy.
type BusySlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// formatEventTime returns the DateTime if set, otherwise the Date (all-day events).
func formatEventTime(edt *api.EventDateTime) string {
	if edt == nil {
		return ""
	}
	if edt.DateTime != "" {
		return edt.DateTime
	}
	return edt.Date
}

// toEventSummary converts a Calendar API event to an EventSummary.
func toEventSummary(event *api.Event) EventSummary {
	s := EventSummary{
		ID:       event.Id,
		Summary:  event.Summary,
		Start:    formatEventTime(event.Start),
		End:      formatEventTime(event.End),
		Location: event.Location,
		Status:   event.Status,
	}
	if event.Organizer != nil {
		s.Organizer = event.Organizer.Email
	}
	return s
}

// toEventDetail converts a Calendar API event to an EventDetail.
func toEventDetail(event *api.Event) EventDetail {
	d := EventDetail{
		ID:          event.Id,
		Summary:     event.Summary,
		Description: event.Description,
		Start:       formatEventTime(event.Start),
		End:         formatEventTime(event.End),
		Location:    event.Location,
		Status:      event.Status,
		HangoutLink: event.HangoutLink,
		HTMLURL:     event.HtmlLink,
		Created:     event.Created,
		Updated:     event.Updated,
	}
	if event.Organizer != nil {
		d.Organizer = event.Organizer.Email
	}
	for _, a := range event.Attendees {
		d.Attendees = append(d.Attendees, AttendeeInfo{
			Email:          a.Email,
			DisplayName:    a.DisplayName,
			ResponseStatus: a.ResponseStatus,
			Organizer:      a.Organizer,
			Self:           a.Self,
		})
	}
	return d
}

// printEventSummaries outputs event summaries as JSON or a formatted text table.
func printEventSummaries(cmd *cobra.Command, events []EventSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(events)
	}

	if len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	lines := make([]string, 0, len(events)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-25s  %-25s  %s", "SUMMARY", "START", "END", "LOCATION"))
	for _, e := range events {
		summary := truncate(e.Summary, 40)
		lines = append(lines, fmt.Sprintf("%-40s  %-25s  %-25s  %s", summary, e.Start, e.End, e.Location))
	}
	cli.PrintText(lines)
	return nil
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// confirmDestructive returns an error if the --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this action is irreversible; re-run with --confirm to proceed")
	}
	return nil
}

// dryRunResult prints a standardised dry-run response and returns nil.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}

// parseAttendees splits a comma-separated string of emails into EventAttendee values.
func parseAttendees(csv string) []*api.EventAttendee {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	attendees := make([]*api.EventAttendee, 0, len(parts))
	for _, email := range parts {
		email = strings.TrimSpace(email)
		if email != "" {
			attendees = append(attendees, &api.EventAttendee{Email: email})
		}
	}
	return attendees
}
