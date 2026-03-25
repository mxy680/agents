package linkedin

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerEventsResponse is the response envelope for GET /voyager/api/events/events.
type voyagerEventsResponse struct {
	Elements []voyagerEventElement `json:"elements"`
	Paging   voyagerPaging         `json:"paging"`
}

type voyagerEventElement struct {
	EntityURN string `json:"entityUrn"`
	Title     string `json:"title"`
	StartAt   int64  `json:"startAt"`
	Location  string `json:"location"`
}

// toEventSummary maps a voyagerEventElement to EventSummary.
func toEventSummary(el voyagerEventElement) EventSummary {
	id := el.EntityURN
	if parts := strings.Split(el.EntityURN, ":"); len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	return EventSummary{
		ID:       id,
		Title:    el.Title,
		StartsAt: el.StartAt,
		Location: el.Location,
	}
}

// newEventsCmd builds the "events" subcommand group.
func newEventsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events",
		Short:   "Interact with LinkedIn events",
		Aliases: []string{"event"},
	}
	cmd.AddCommand(newEventsListCmd(factory))
	cmd.AddCommand(newEventsGetCmd(factory))
	return cmd
}

// newEventsListCmd builds the "events list" command.
func newEventsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events you are attending or have been invited to",
		RunE:  makeRunEventsList(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of events to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newEventsGetCmd builds the "events get" command.
func newEventsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an event by ID",
		RunE:  makeRunEventsGet(factory),
	}
	cmd.Flags().String("event-id", "", "Event ID (required)")
	_ = cmd.MarkFlagRequired("event-id")
	return cmd
}

func makeRunEventsList(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunEventsGet(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

// printEventSummaries outputs event summaries as JSON or text.
func printEventSummaries(cmd *cobra.Command, events []EventSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(events)
	}
	if len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}
	lines := make([]string, 0, len(events)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-16s  %-30s", "ID", "TITLE", "STARTS AT", "LOCATION"))
	for _, e := range events {
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-16s  %-30s",
			truncate(e.ID, 20),
			truncate(e.Title, 40),
			formatTimestamp(e.StartsAt),
			truncate(e.Location, 30),
		))
	}
	cli.PrintText(lines)
	return nil
}

// printEventDetail outputs a single event as JSON or formatted text block.
func printEventDetail(cmd *cobra.Command, e EventSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(e)
	}
	lines := []string{
		fmt.Sprintf("ID:       %s", e.ID),
		fmt.Sprintf("Title:    %s", e.Title),
		fmt.Sprintf("Starts:   %s", formatTimestamp(e.StartsAt)),
		fmt.Sprintf("Location: %s", e.Location),
	}
	cli.PrintText(lines)
	return nil
}
