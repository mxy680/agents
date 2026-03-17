package calendar

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/calendar/v3"
)

func newEventsListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List calendar events",
		RunE:  makeRunEventsList(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("query", "", "Free-text search terms")
	cmd.Flags().String("time-min", "", "Lower bound (RFC3339) for event start")
	cmd.Flags().String("time-max", "", "Upper bound (RFC3339) for event start")
	cmd.Flags().Int("limit", 20, "Maximum number of events to return")
	cmd.Flags().String("page-token", "", "Token for next page of results")
	cmd.Flags().Bool("single-events", false, "Expand recurring events into instances")
	cmd.Flags().String("order-by", "", "Sort order: startTime or updated (requires --single-events for startTime)")
	return cmd
}

func makeRunEventsList(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		query, _ := cmd.Flags().GetString("query")
		timeMin, _ := cmd.Flags().GetString("time-min")
		timeMax, _ := cmd.Flags().GetString("time-max")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")
		singleEvents, _ := cmd.Flags().GetBool("single-events")
		orderBy, _ := cmd.Flags().GetString("order-by")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := svc.Events.List(calendarID).MaxResults(int64(limit)).SingleEvents(singleEvents)
		if query != "" {
			req = req.Q(query)
		}
		if timeMin != "" {
			req = req.TimeMin(timeMin)
		}
		if timeMax != "" {
			req = req.TimeMax(timeMax)
		}
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		if orderBy != "" {
			req = req.OrderBy(orderBy)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing events: %w", err)
		}

		summaries := make([]EventSummary, 0, len(resp.Items))
		for _, item := range resp.Items {
			summaries = append(summaries, toEventSummary(item))
		}
		return printEventSummaries(cmd, summaries)
	}
}

func newEventsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a calendar event by ID",
		RunE:  makeRunEventsGet(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("event-id", "", "Event ID (required)")
	_ = cmd.MarkFlagRequired("event-id")
	return cmd
}

func makeRunEventsGet(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		event, err := svc.Events.Get(calendarID, eventID).Do()
		if err != nil {
			return fmt.Errorf("getting event %s: %w", eventID, err)
		}

		detail := toEventDetail(event)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", detail.ID),
			fmt.Sprintf("Summary:     %s", detail.Summary),
			fmt.Sprintf("Start:       %s", detail.Start),
			fmt.Sprintf("End:         %s", detail.End),
			fmt.Sprintf("Location:    %s", detail.Location),
			fmt.Sprintf("Status:      %s", detail.Status),
			fmt.Sprintf("Organizer:   %s", detail.Organizer),
			fmt.Sprintf("Description: %s", detail.Description),
		}
		if detail.HangoutLink != "" {
			lines = append(lines, fmt.Sprintf("Hangout:     %s", detail.HangoutLink))
		}
		for _, a := range detail.Attendees {
			lines = append(lines, fmt.Sprintf("Attendee:    %s (%s)", a.Email, a.ResponseStatus))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newEventsCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a calendar event",
		RunE:  makeRunEventsCreate(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("summary", "", "Event title (required)")
	cmd.Flags().String("start", "", "Start time RFC3339 or YYYY-MM-DD for all-day (required)")
	cmd.Flags().String("end", "", "End time RFC3339 or YYYY-MM-DD for all-day (required)")
	cmd.Flags().String("description", "", "Event description")
	cmd.Flags().String("location", "", "Event location")
	cmd.Flags().String("attendees", "", "Comma-separated attendee emails")
	cmd.Flags().String("timezone", "", "Timezone (e.g. America/New_York)")
	cmd.Flags().Bool("all-day", false, "Create an all-day event (use YYYY-MM-DD for start/end)")
	_ = cmd.MarkFlagRequired("summary")
	_ = cmd.MarkFlagRequired("start")
	_ = cmd.MarkFlagRequired("end")
	return cmd
}

func makeRunEventsCreate(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		summary, _ := cmd.Flags().GetString("summary")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		description, _ := cmd.Flags().GetString("description")
		location, _ := cmd.Flags().GetString("location")
		attendees, _ := cmd.Flags().GetString("attendees")
		timezone, _ := cmd.Flags().GetString("timezone")
		allDay, _ := cmd.Flags().GetBool("all-day")

		event := &api.Event{
			Summary:     summary,
			Description: description,
			Location:    location,
			Attendees:   parseAttendees(attendees),
		}

		if allDay {
			event.Start = &api.EventDateTime{Date: start}
			event.End = &api.EventDateTime{Date: end}
		} else {
			event.Start = &api.EventDateTime{DateTime: start, TimeZone: timezone}
			event.End = &api.EventDateTime{DateTime: end, TimeZone: timezone}
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create event %q on calendar %s", summary, calendarID), map[string]any{
				"action":     "create",
				"calendarId": calendarID,
				"summary":    summary,
				"start":      start,
				"end":        end,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		created, err := svc.Events.Insert(calendarID, event).Do()
		if err != nil {
			return fmt.Errorf("creating event: %w", err)
		}

		detail := toEventDetail(created)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created event: %s (%s)\n", detail.Summary, detail.ID)
		return nil
	}
}

func newEventsQuickAddCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quick-add",
		Short: "Quick-add an event from natural language text",
		RunE:  makeRunEventsQuickAdd(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("text", "", "Natural language event description (required)")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunEventsQuickAdd(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		text, _ := cmd.Flags().GetString("text")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would quick-add event %q on calendar %s", text, calendarID), map[string]any{
				"action":     "quick-add",
				"calendarId": calendarID,
				"text":       text,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		created, err := svc.Events.QuickAdd(calendarID, text).Do()
		if err != nil {
			return fmt.Errorf("quick-add event: %w", err)
		}

		detail := toEventDetail(created)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Created event: %s (%s)\n", detail.Summary, detail.ID)
		return nil
	}
}

func newEventsUpdateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a calendar event",
		RunE:  makeRunEventsUpdate(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("event-id", "", "Event ID (required)")
	cmd.Flags().String("summary", "", "New event title")
	cmd.Flags().String("start", "", "New start time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("end", "", "New end time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().String("description", "", "New event description")
	cmd.Flags().String("location", "", "New event location")
	cmd.Flags().String("attendees", "", "Comma-separated attendee emails (replaces existing)")
	_ = cmd.MarkFlagRequired("event-id")
	return cmd
}

func makeRunEventsUpdate(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")
		summary, _ := cmd.Flags().GetString("summary")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		description, _ := cmd.Flags().GetString("description")
		location, _ := cmd.Flags().GetString("location")
		attendees, _ := cmd.Flags().GetString("attendees")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update event %s on calendar %s", eventID, calendarID), map[string]any{
				"action":     "update",
				"calendarId": calendarID,
				"eventId":    eventID,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		// Fetch current event to merge changes
		event, err := svc.Events.Get(calendarID, eventID).Do()
		if err != nil {
			return fmt.Errorf("getting event %s for update: %w", eventID, err)
		}

		if summary != "" {
			event.Summary = summary
		}
		if description != "" {
			event.Description = description
		}
		if location != "" {
			event.Location = location
		}
		if start != "" {
			if event.Start == nil {
				event.Start = &api.EventDateTime{}
			}
			if len(start) == 10 { // YYYY-MM-DD
				event.Start.Date = start
				event.Start.DateTime = ""
			} else {
				event.Start.DateTime = start
				event.Start.Date = ""
			}
		}
		if end != "" {
			if event.End == nil {
				event.End = &api.EventDateTime{}
			}
			if len(end) == 10 {
				event.End.Date = end
				event.End.DateTime = ""
			} else {
				event.End.DateTime = end
				event.End.Date = ""
			}
		}
		if cmd.Flags().Changed("attendees") {
			event.Attendees = parseAttendees(attendees)
		}

		updated, err := svc.Events.Update(calendarID, eventID, event).Do()
		if err != nil {
			return fmt.Errorf("updating event %s: %w", eventID, err)
		}

		detail := toEventDetail(updated)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Updated event: %s (%s)\n", detail.Summary, detail.ID)
		return nil
	}
}

func newEventsDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a calendar event",
		RunE:  makeRunEventsDelete(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("event-id", "", "Event ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("event-id")
	return cmd
}

func makeRunEventsDelete(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete event %s from calendar %s", eventID, calendarID), map[string]any{
				"action":     "delete",
				"calendarId": calendarID,
				"eventId":    eventID,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		if err := svc.Events.Delete(calendarID, eventID).Do(); err != nil {
			return fmt.Errorf("deleting event %s: %w", eventID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "eventId": eventID})
		}
		fmt.Printf("Deleted event: %s\n", eventID)
		return nil
	}
}

func newEventsMoveCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move an event to a different calendar",
		RunE:  makeRunEventsMove(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Source calendar ID")
	cmd.Flags().String("event-id", "", "Event ID (required)")
	cmd.Flags().String("destination", "", "Destination calendar ID (required)")
	_ = cmd.MarkFlagRequired("event-id")
	_ = cmd.MarkFlagRequired("destination")
	return cmd
}

func makeRunEventsMove(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")
		destination, _ := cmd.Flags().GetString("destination")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		moved, err := svc.Events.Move(calendarID, eventID, destination).Do()
		if err != nil {
			return fmt.Errorf("moving event %s: %w", eventID, err)
		}

		detail := toEventDetail(moved)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Moved event: %s → calendar %s\n", detail.Summary, destination)
		return nil
	}
}

func newEventsInstancesCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "List instances of a recurring event",
		RunE:  makeRunEventsInstances(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("event-id", "", "Recurring event ID (required)")
	cmd.Flags().String("time-min", "", "Lower bound (RFC3339)")
	cmd.Flags().String("time-max", "", "Upper bound (RFC3339)")
	cmd.Flags().Int("limit", 20, "Maximum number of instances to return")
	cmd.Flags().String("page-token", "", "Token for next page of results")
	_ = cmd.MarkFlagRequired("event-id")
	return cmd
}

func makeRunEventsInstances(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")
		timeMin, _ := cmd.Flags().GetString("time-min")
		timeMax, _ := cmd.Flags().GetString("time-max")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := svc.Events.Instances(calendarID, eventID).MaxResults(int64(limit))
		if timeMin != "" {
			req = req.TimeMin(timeMin)
		}
		if timeMax != "" {
			req = req.TimeMax(timeMax)
		}
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing instances of event %s: %w", eventID, err)
		}

		summaries := make([]EventSummary, 0, len(resp.Items))
		for _, item := range resp.Items {
			summaries = append(summaries, toEventSummary(item))
		}
		return printEventSummaries(cmd, summaries)
	}
}
