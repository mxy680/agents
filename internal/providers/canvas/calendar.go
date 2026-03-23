package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newCalendarCmd returns the parent "calendar" command with all subcommands attached.
func newCalendarCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "calendar",
		Short:   "Manage Canvas calendar events",
		Aliases: []string{"cal"},
	}

	cmd.AddCommand(newCalendarListCmd(factory))
	cmd.AddCommand(newCalendarGetCmd(factory))
	cmd.AddCommand(newCalendarCreateCmd(factory))
	cmd.AddCommand(newCalendarUpdateCmd(factory))
	cmd.AddCommand(newCalendarDeleteCmd(factory))

	return cmd
}

func newCalendarListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List calendar events",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			eventType, _ := cmd.Flags().GetString("type")
			startDate, _ := cmd.Flags().GetString("start-date")
			endDate, _ := cmd.Flags().GetString("end-date")
			contextCodes, _ := cmd.Flags().GetStringSlice("context-codes")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if eventType != "" {
				params.Set("type", eventType)
			}
			if startDate != "" {
				params.Set("start_date", startDate)
			}
			if endDate != "" {
				params.Set("end_date", endDate)
			}
			for _, code := range contextCodes {
				code = strings.TrimSpace(code)
				if code != "" {
					params.Add("context_codes[]", code)
				}
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/calendar_events", params)
			if err != nil {
				return err
			}

			var events []CalendarEventSummary
			if err := json.Unmarshal(data, &events); err != nil {
				return fmt.Errorf("parse calendar events: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(events)
			}

			if len(events) == 0 {
				fmt.Println("No calendar events found.")
				return nil
			}
			for _, e := range events {
				fmt.Printf("%-6d  %-25s  %-15s  %s\n", e.ID, e.StartAt, e.ContextCode, truncate(e.Title, 50))
			}
			return nil
		},
	}

	cmd.Flags().String("type", "", "Filter by type: event or assignment")
	cmd.Flags().String("start-date", "", "Start date in RFC3339 format")
	cmd.Flags().String("end-date", "", "End date in RFC3339 format")
	cmd.Flags().StringSlice("context-codes", nil, "Context codes to filter by (e.g. course_123)")
	cmd.Flags().Int("limit", 0, "Maximum number of events to return")
	return cmd
}

func newCalendarGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific calendar event",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			eventID, _ := cmd.Flags().GetString("event-id")
			if eventID == "" {
				return fmt.Errorf("--event-id is required")
			}

			data, err := client.Get(ctx, "/calendar_events/"+eventID, nil)
			if err != nil {
				return err
			}

			var event CalendarEventSummary
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("parse calendar event: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(event)
			}

			fmt.Printf("ID:           %d\n", event.ID)
			fmt.Printf("Title:        %s\n", event.Title)
			fmt.Printf("Context:      %s\n", event.ContextCode)
			fmt.Printf("State:        %s\n", event.WorkflowState)
			if event.StartAt != "" {
				fmt.Printf("Start:        %s\n", event.StartAt)
			}
			if event.EndAt != "" {
				fmt.Printf("End:          %s\n", event.EndAt)
			}
			if event.LocationName != "" {
				fmt.Printf("Location:     %s\n", event.LocationName)
			}
			if event.AllDay {
				fmt.Println("All Day:      yes")
			}
			if event.Description != "" {
				fmt.Printf("Description:  %s\n", truncate(event.Description, 200))
			}
			return nil
		},
	}

	cmd.Flags().String("event-id", "", "Canvas calendar event ID (required)")
	return cmd
}

func newCalendarCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new calendar event",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			contextCode, _ := cmd.Flags().GetString("context-code")
			title, _ := cmd.Flags().GetString("title")
			startAt, _ := cmd.Flags().GetString("start-at")
			endAt, _ := cmd.Flags().GetString("end-at")
			if contextCode == "" {
				return fmt.Errorf("--context-code is required")
			}
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if startAt == "" {
				return fmt.Errorf("--start-at is required")
			}
			if endAt == "" {
				return fmt.Errorf("--end-at is required")
			}

			description, _ := cmd.Flags().GetString("description")
			locationName, _ := cmd.Flags().GetString("location-name")
			allDay, _ := cmd.Flags().GetBool("all-day")

			eventBody := map[string]any{
				"context_code": contextCode,
				"title":        title,
				"start_at":     startAt,
				"end_at":       endAt,
			}
			if description != "" {
				eventBody["description"] = description
			}
			if locationName != "" {
				eventBody["location_name"] = locationName
			}
			if allDay {
				eventBody["all_day"] = true
			}

			body := map[string]any{"calendar_event": eventBody}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create calendar event %q in %s", title, contextCode), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/calendar_events", body)
			if err != nil {
				return err
			}

			var event CalendarEventSummary
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("parse calendar event: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(event)
			}

			fmt.Printf("Created calendar event %d: %s\n", event.ID, event.Title)
			return nil
		},
	}

	cmd.Flags().String("context-code", "", "Context code for the event, e.g. course_123 (required)")
	cmd.Flags().String("title", "", "Event title (required)")
	cmd.Flags().String("start-at", "", "Start time in RFC3339 format (required)")
	cmd.Flags().String("end-at", "", "End time in RFC3339 format (required)")
	cmd.Flags().String("description", "", "Event description")
	cmd.Flags().String("location-name", "", "Location name")
	cmd.Flags().Bool("all-day", false, "Mark as an all-day event")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newCalendarUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing calendar event",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			eventID, _ := cmd.Flags().GetString("event-id")
			if eventID == "" {
				return fmt.Errorf("--event-id is required")
			}

			eventBody := map[string]any{}
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				eventBody["title"] = v
			}
			if cmd.Flags().Changed("start-at") {
				v, _ := cmd.Flags().GetString("start-at")
				eventBody["start_at"] = v
			}
			if cmd.Flags().Changed("end-at") {
				v, _ := cmd.Flags().GetString("end-at")
				eventBody["end_at"] = v
			}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				eventBody["description"] = v
			}

			body := map[string]any{"calendar_event": eventBody}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update calendar event %s", eventID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/calendar_events/"+eventID, body)
			if err != nil {
				return err
			}

			var event CalendarEventSummary
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("parse calendar event: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(event)
			}

			fmt.Printf("Updated calendar event %d: %s\n", event.ID, event.Title)
			return nil
		},
	}

	cmd.Flags().String("event-id", "", "Canvas calendar event ID (required)")
	cmd.Flags().String("title", "", "New event title")
	cmd.Flags().String("start-at", "", "New start time in RFC3339 format")
	cmd.Flags().String("end-at", "", "New end time in RFC3339 format")
	cmd.Flags().String("description", "", "New event description")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newCalendarDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a calendar event",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			eventID, _ := cmd.Flags().GetString("event-id")
			if eventID == "" {
				return fmt.Errorf("--event-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the calendar event"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete calendar event %s", eventID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/calendar_events/"+eventID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "event_id": eventID})
			}
			fmt.Printf("Calendar event %s deleted.\n", eventID)
			return nil
		},
	}

	cmd.Flags().String("event-id", "", "Canvas calendar event ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
