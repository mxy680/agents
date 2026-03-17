package calendar

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/calendar/v3"
)

func newCalendarsListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List calendars",
		RunE:  makeRunCalendarsList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of calendars to return")
	cmd.Flags().String("page-token", "", "Token for next page of results")
	cmd.Flags().Bool("show-hidden", false, "Include hidden calendars")
	return cmd
}

func makeRunCalendarsList(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")
		showHidden, _ := cmd.Flags().GetBool("show-hidden")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := svc.CalendarList.List().MaxResults(int64(limit)).ShowHidden(showHidden)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing calendars: %w", err)
		}

		summaries := make([]CalendarSummary, 0, len(resp.Items))
		for _, item := range resp.Items {
			summaries = append(summaries, CalendarSummary{
				ID:          item.Id,
				Summary:     item.Summary,
				Description: item.Description,
				TimeZone:    item.TimeZone,
				Primary:     item.Primary,
				AccessRole:  item.AccessRole,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No calendars found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-20s  %s", "ID", "SUMMARY", "TIMEZONE", "ACCESS"))
		for _, c := range summaries {
			id := truncate(c.ID, 40)
			summary := truncate(c.Summary, 30)
			lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-20s  %s", id, summary, c.TimeZone, c.AccessRole))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newCalendarsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get calendar details",
		RunE:  makeRunCalendarsGet(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	return cmd
}

func makeRunCalendarsGet(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		entry, err := svc.CalendarList.Get(calendarID).Do()
		if err != nil {
			return fmt.Errorf("getting calendar %s: %w", calendarID, err)
		}

		summary := CalendarSummary{
			ID:          entry.Id,
			Summary:     entry.Summary,
			Description: entry.Description,
			TimeZone:    entry.TimeZone,
			Primary:     entry.Primary,
			AccessRole:  entry.AccessRole,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", summary.ID),
			fmt.Sprintf("Summary:     %s", summary.Summary),
			fmt.Sprintf("Description: %s", summary.Description),
			fmt.Sprintf("TimeZone:    %s", summary.TimeZone),
			fmt.Sprintf("Primary:     %v", summary.Primary),
			fmt.Sprintf("AccessRole:  %s", summary.AccessRole),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newCalendarsCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new calendar",
		RunE:  makeRunCalendarsCreate(factory),
	}
	cmd.Flags().String("summary", "", "Calendar name (required)")
	cmd.Flags().String("description", "", "Calendar description")
	cmd.Flags().String("timezone", "", "Calendar timezone (e.g. America/New_York)")
	_ = cmd.MarkFlagRequired("summary")
	return cmd
}

func makeRunCalendarsCreate(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		summary, _ := cmd.Flags().GetString("summary")
		description, _ := cmd.Flags().GetString("description")
		timezone, _ := cmd.Flags().GetString("timezone")

		cal := &api.Calendar{
			Summary:     summary,
			Description: description,
			TimeZone:    timezone,
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create calendar %q", summary), map[string]any{
				"action":  "create",
				"summary": summary,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		created, err := svc.Calendars.Insert(cal).Do()
		if err != nil {
			return fmt.Errorf("creating calendar: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(CalendarSummary{
				ID:          created.Id,
				Summary:     created.Summary,
				Description: created.Description,
				TimeZone:    created.TimeZone,
			})
		}
		fmt.Printf("Created calendar: %s (%s)\n", created.Summary, created.Id)
		return nil
	}
}

func newCalendarsUpdateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a calendar",
		RunE:  makeRunCalendarsUpdate(factory),
	}
	cmd.Flags().String("calendar-id", "primary", "Calendar ID")
	cmd.Flags().String("summary", "", "New calendar name")
	cmd.Flags().String("description", "", "New calendar description")
	cmd.Flags().String("timezone", "", "New calendar timezone")
	return cmd
}

func makeRunCalendarsUpdate(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")
		summary, _ := cmd.Flags().GetString("summary")
		description, _ := cmd.Flags().GetString("description")
		timezone, _ := cmd.Flags().GetString("timezone")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update calendar %s", calendarID), map[string]any{
				"action":     "update",
				"calendarId": calendarID,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		// Fetch current calendar to merge changes
		cal, err := svc.Calendars.Get(calendarID).Do()
		if err != nil {
			return fmt.Errorf("getting calendar %s for update: %w", calendarID, err)
		}

		if summary != "" {
			cal.Summary = summary
		}
		if description != "" {
			cal.Description = description
		}
		if timezone != "" {
			cal.TimeZone = timezone
		}

		updated, err := svc.Calendars.Update(calendarID, cal).Do()
		if err != nil {
			return fmt.Errorf("updating calendar %s: %w", calendarID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(CalendarSummary{
				ID:          updated.Id,
				Summary:     updated.Summary,
				Description: updated.Description,
				TimeZone:    updated.TimeZone,
			})
		}
		fmt.Printf("Updated calendar: %s (%s)\n", updated.Summary, updated.Id)
		return nil
	}
}

func newCalendarsDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a calendar",
		RunE:  makeRunCalendarsDelete(factory),
	}
	cmd.Flags().String("calendar-id", "", "Calendar ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("calendar-id")
	return cmd
}

func makeRunCalendarsDelete(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarID, _ := cmd.Flags().GetString("calendar-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete calendar %s", calendarID), map[string]any{
				"action":     "delete",
				"calendarId": calendarID,
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

		if err := svc.Calendars.Delete(calendarID).Do(); err != nil {
			return fmt.Errorf("deleting calendar %s: %w", calendarID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "calendarId": calendarID})
		}
		fmt.Printf("Deleted calendar: %s\n", calendarID)
		return nil
	}
}
