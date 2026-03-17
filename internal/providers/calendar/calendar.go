package calendar

import (
	"context"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
	api "google.golang.org/api/calendar/v3"
)

// ServiceFactory is a function that creates a Calendar API service.
type ServiceFactory func(ctx context.Context) (*api.Service, error)

// Provider implements the Google Calendar integration.
type Provider struct {
	// ServiceFactory creates the Calendar API service. Defaults to auth.NewCalendarService.
	// Override in tests to inject a mock service pointing at a test server.
	ServiceFactory ServiceFactory
}

// New creates a new Calendar provider using the real Calendar API.
func New() *Provider {
	return &Provider{
		ServiceFactory: auth.NewCalendarService,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "calendar"
}

// RegisterCommands adds all Calendar subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	calendarCmd := &cobra.Command{
		Use:   "calendar",
		Short: "Interact with Google Calendar",
		Long:  "List, create, update, and delete calendar events via the Google Calendar API.",
	}

	eventsCmd := &cobra.Command{
		Use:     "events",
		Short:   "Manage calendar events",
		Aliases: []string{"event", "ev"},
	}
	eventsCmd.AddCommand(newEventsListCmd(p.ServiceFactory))
	eventsCmd.AddCommand(newEventsGetCmd(p.ServiceFactory))
	eventsCmd.AddCommand(newEventsCreateCmd(p.ServiceFactory))
	eventsCmd.AddCommand(newEventsQuickAddCmd(p.ServiceFactory))
	eventsCmd.AddCommand(newEventsUpdateCmd(p.ServiceFactory))
	eventsCmd.AddCommand(newEventsDeleteCmd(p.ServiceFactory))
	eventsCmd.AddCommand(newEventsMoveCmd(p.ServiceFactory))
	eventsCmd.AddCommand(newEventsInstancesCmd(p.ServiceFactory))
	calendarCmd.AddCommand(eventsCmd)

	calendarsCmd := &cobra.Command{
		Use:     "calendars",
		Short:   "Manage calendars",
		Aliases: []string{"cal"},
	}
	calendarsCmd.AddCommand(newCalendarsListCmd(p.ServiceFactory))
	calendarsCmd.AddCommand(newCalendarsGetCmd(p.ServiceFactory))
	calendarsCmd.AddCommand(newCalendarsCreateCmd(p.ServiceFactory))
	calendarsCmd.AddCommand(newCalendarsUpdateCmd(p.ServiceFactory))
	calendarsCmd.AddCommand(newCalendarsDeleteCmd(p.ServiceFactory))
	calendarCmd.AddCommand(calendarsCmd)

	freebusyCmd := &cobra.Command{
		Use:     "freebusy",
		Short:   "Query free/busy information",
		Aliases: []string{"fb"},
	}
	freebusyCmd.AddCommand(newFreebusyQueryCmd(p.ServiceFactory))
	calendarCmd.AddCommand(freebusyCmd)

	parent.AddCommand(calendarCmd)
}
