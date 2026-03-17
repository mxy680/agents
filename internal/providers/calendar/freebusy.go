package calendar

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/calendar/v3"
)

func newFreebusyQueryCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query free/busy information for calendars",
		RunE:  makeRunFreebusyQuery(factory),
	}
	cmd.Flags().String("calendar-ids", "primary", "Comma-separated calendar IDs")
	cmd.Flags().String("time-min", "", "Start of time range (RFC3339, required)")
	cmd.Flags().String("time-max", "", "End of time range (RFC3339, required)")
	_ = cmd.MarkFlagRequired("time-min")
	_ = cmd.MarkFlagRequired("time-max")
	return cmd
}

func makeRunFreebusyQuery(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		calendarIDs, _ := cmd.Flags().GetString("calendar-ids")
		timeMin, _ := cmd.Flags().GetString("time-min")
		timeMax, _ := cmd.Flags().GetString("time-max")

		ids := strings.Split(calendarIDs, ",")
		items := make([]*api.FreeBusyRequestItem, 0, len(ids))
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id != "" {
				items = append(items, &api.FreeBusyRequestItem{Id: id})
			}
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := &api.FreeBusyRequest{
			TimeMin: timeMin,
			TimeMax: timeMax,
			Items:   items,
		}

		resp, err := svc.Freebusy.Query(req).Do()
		if err != nil {
			return fmt.Errorf("querying free/busy: %w", err)
		}

		results := make([]FreeBusyResult, 0, len(resp.Calendars))
		for calID, cal := range resp.Calendars {
			result := FreeBusyResult{CalendarID: calID}
			for _, busy := range cal.Busy {
				result.Busy = append(result.Busy, BusySlot{
					Start: busy.Start,
					End:   busy.End,
				})
			}
			results = append(results, result)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(results)
		}

		if len(results) == 0 {
			fmt.Println("No free/busy data found.")
			return nil
		}

		for _, r := range results {
			fmt.Printf("Calendar: %s\n", r.CalendarID)
			if len(r.Busy) == 0 {
				fmt.Println("  No busy times in range.")
			}
			for _, b := range r.Busy {
				fmt.Printf("  Busy: %s → %s\n", b.Start, b.End)
			}
		}
		return nil
	}
}
