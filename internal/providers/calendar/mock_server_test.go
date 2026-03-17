package calendar

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
	api "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// withEventsMock registers all event-related mock handlers on mux.
func withEventsMock(mux *http.ServeMux) {
	// events.list
	mux.HandleFunc("/calendars/primary/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// events.insert
			var event map[string]any
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &event)
			resp := map[string]any{
				"id":      "ev-created1",
				"summary": event["summary"],
				"status":  "confirmed",
				"start":   event["start"],
				"end":     event["end"],
				"created": "2026-03-16T10:00:00Z",
				"updated": "2026-03-16T10:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// events.list
		resp := map[string]any{
			"items": []map[string]any{
				{
					"id":      "ev1",
					"summary": "Team Standup",
					"start":   map[string]string{"dateTime": "2026-03-16T09:00:00-04:00"},
					"end":     map[string]string{"dateTime": "2026-03-16T09:30:00-04:00"},
					"status":  "confirmed",
					"organizer": map[string]any{
						"email": "alice@example.com",
					},
				},
				{
					"id":      "ev2",
					"summary": "Lunch",
					"start":   map[string]string{"date": "2026-03-16"},
					"end":     map[string]string{"date": "2026-03-17"},
					"status":  "confirmed",
					"location": "Cafeteria",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// events.quickAdd
	mux.HandleFunc("/calendars/primary/events/quickAdd", func(w http.ResponseWriter, r *http.Request) {
		text := r.URL.Query().Get("text")
		resp := map[string]any{
			"id":      "ev-quick1",
			"summary": text,
			"status":  "confirmed",
			"start":   map[string]string{"dateTime": "2026-03-17T15:00:00-04:00"},
			"end":     map[string]string{"dateTime": "2026-03-17T16:00:00-04:00"},
			"created": "2026-03-16T10:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// events.get, events.update, events.delete for ev1
	mux.HandleFunc("/calendars/primary/events/ev1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPut {
			// events.update
			var event map[string]any
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &event)
			resp := map[string]any{
				"id":      "ev1",
				"summary": event["summary"],
				"status":  "confirmed",
				"start":   event["start"],
				"end":     event["end"],
				"updated": "2026-03-16T12:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// events.get
		resp := map[string]any{
			"id":          "ev1",
			"summary":     "Team Standup",
			"description": "Daily standup meeting",
			"start":       map[string]string{"dateTime": "2026-03-16T09:00:00-04:00"},
			"end":         map[string]string{"dateTime": "2026-03-16T09:30:00-04:00"},
			"status":      "confirmed",
			"location":    "Room 101",
			"organizer":   map[string]any{"email": "alice@example.com"},
			"attendees": []map[string]any{
				{"email": "bob@example.com", "responseStatus": "accepted"},
				{"email": "charlie@example.com", "responseStatus": "tentative"},
			},
			"htmlLink": "https://calendar.google.com/event?eid=ev1",
			"created":  "2026-03-01T10:00:00Z",
			"updated":  "2026-03-15T10:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// events.move for ev1
	mux.HandleFunc("/calendars/primary/events/ev1/move", func(w http.ResponseWriter, r *http.Request) {
		dest := r.URL.Query().Get("destination")
		resp := map[string]any{
			"id":      "ev1",
			"summary": "Team Standup",
			"status":  "confirmed",
			"start":   map[string]string{"dateTime": "2026-03-16T09:00:00-04:00"},
			"end":     map[string]string{"dateTime": "2026-03-16T09:30:00-04:00"},
		}
		_ = dest
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// events.instances for ev1
	mux.HandleFunc("/calendars/primary/events/ev1/instances", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items": []map[string]any{
				{
					"id":      "ev1_20260316",
					"summary": "Team Standup",
					"start":   map[string]string{"dateTime": "2026-03-16T09:00:00-04:00"},
					"end":     map[string]string{"dateTime": "2026-03-16T09:30:00-04:00"},
					"status":  "confirmed",
				},
				{
					"id":      "ev1_20260317",
					"summary": "Team Standup",
					"start":   map[string]string{"dateTime": "2026-03-17T09:00:00-04:00"},
					"end":     map[string]string{"dateTime": "2026-03-17T09:30:00-04:00"},
					"status":  "confirmed",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withCalendarsMock registers all calendar-list and calendar-related mock handlers on mux.
func withCalendarsMock(mux *http.ServeMux) {
	// calendarList.list
	mux.HandleFunc("/users/me/calendarList", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items": []map[string]any{
				{
					"id":         "primary",
					"summary":    "My Calendar",
					"timeZone":   "America/New_York",
					"primary":    true,
					"accessRole": "owner",
				},
				{
					"id":         "work@group.calendar.google.com",
					"summary":    "Work Calendar",
					"timeZone":   "America/New_York",
					"accessRole": "writer",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// calendarList.get
	mux.HandleFunc("/users/me/calendarList/primary", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id":          "primary",
			"summary":     "My Calendar",
			"description": "Main calendar",
			"timeZone":    "America/New_York",
			"primary":     true,
			"accessRole":  "owner",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// calendars.insert (POST to /calendars)
	mux.HandleFunc("/calendars", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var cal map[string]any
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &cal)
			resp := map[string]any{
				"id":       "new-cal-123",
				"summary":  cal["summary"],
				"timeZone": cal["timeZone"],
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	})

	// calendars.get, calendars.update, calendars.delete for primary
	mux.HandleFunc("/calendars/primary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPut {
			var cal map[string]any
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &cal)
			resp := map[string]any{
				"id":       "primary",
				"summary":  cal["summary"],
				"timeZone": cal["timeZone"],
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// GET
		resp := map[string]any{
			"id":          "primary",
			"summary":     "My Calendar",
			"description": "Main calendar",
			"timeZone":    "America/New_York",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// calendars.get, calendars.delete for work calendar
	mux.HandleFunc("/calendars/work-cal-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"id":       "work-cal-1",
			"summary":  "Work Calendar",
			"timeZone": "America/New_York",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withFreebusyMock registers free/busy mock handlers on mux.
func withFreebusyMock(mux *http.ServeMux) {
	mux.HandleFunc("/freeBusy", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)

		resp := map[string]any{
			"calendars": map[string]any{
				"primary": map[string]any{
					"busy": []map[string]string{
						{"start": "2026-03-16T09:00:00-04:00", "end": "2026-03-16T09:30:00-04:00"},
						{"start": "2026-03-16T14:00:00-04:00", "end": "2026-03-16T15:00:00-04:00"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates an httptest server with all Calendar mock handlers.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withEventsMock(mux)
	withCalendarsMock(mux)
	withFreebusyMock(mux)
	return httptest.NewServer(mux)
}

// newTestServiceFactory returns a ServiceFactory that creates a *calendar.Service
// backed by the given httptest server, bypassing OAuth entirely.
func newTestServiceFactory(server *httptest.Server) ServiceFactory {
	return func(ctx context.Context) (*api.Service, error) {
		return api.NewService(ctx,
			option.WithoutAuthentication(),
			option.WithEndpoint(server.URL+"/"),
			option.WithHTTPClient(server.Client()),
		)
	}
}

// captureStdout runs f with os.Stdout redirected to a pipe and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	buf := make([]byte, 65536)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// buildTestEventsCmd creates an `events` subcommand tree for use in tests.
func buildTestEventsCmd(factory ServiceFactory) *cobra.Command {
	eventsCmd := &cobra.Command{Use: "events", Aliases: []string{"event", "ev"}}
	eventsCmd.AddCommand(newEventsListCmd(factory))
	eventsCmd.AddCommand(newEventsGetCmd(factory))
	eventsCmd.AddCommand(newEventsCreateCmd(factory))
	eventsCmd.AddCommand(newEventsQuickAddCmd(factory))
	eventsCmd.AddCommand(newEventsUpdateCmd(factory))
	eventsCmd.AddCommand(newEventsDeleteCmd(factory))
	eventsCmd.AddCommand(newEventsMoveCmd(factory))
	eventsCmd.AddCommand(newEventsInstancesCmd(factory))
	return eventsCmd
}

// buildTestCalendarsCmd creates a `calendars` subcommand tree for use in tests.
func buildTestCalendarsCmd(factory ServiceFactory) *cobra.Command {
	calendarsCmd := &cobra.Command{Use: "calendars", Aliases: []string{"cal"}}
	calendarsCmd.AddCommand(newCalendarsListCmd(factory))
	calendarsCmd.AddCommand(newCalendarsGetCmd(factory))
	calendarsCmd.AddCommand(newCalendarsCreateCmd(factory))
	calendarsCmd.AddCommand(newCalendarsUpdateCmd(factory))
	calendarsCmd.AddCommand(newCalendarsDeleteCmd(factory))
	return calendarsCmd
}

// buildTestFreebusyCmd creates a `freebusy` subcommand tree for use in tests.
func buildTestFreebusyCmd(factory ServiceFactory) *cobra.Command {
	freebusyCmd := &cobra.Command{Use: "freebusy", Aliases: []string{"fb"}}
	freebusyCmd.AddCommand(newFreebusyQueryCmd(factory))
	return freebusyCmd
}
