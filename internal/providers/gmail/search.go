package gmail

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSearchCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Gmail using Gmail query syntax",
		RunE:  makeRunSearch(factory),
	}
	cmd.Flags().String("query", "", "Gmail search query (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of results")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunSearch(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Users.Messages.List("me").Q(query).MaxResults(int64(limit)).Do()
		if err != nil {
			return fmt.Errorf("searching messages: %w", err)
		}

		summaries, err := fetchSummaries(ctx, svc, resp.Messages)
		if err != nil {
			return err
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No messages found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", "FROM", "SUBJECT", "DATE"))
		for _, s := range summaries {
			from := truncate(s.From, 20)
			subject := truncate(s.Subject, 40)
			lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", from, subject, s.Date))
		}
		cli.PrintText(lines)
		return nil
	}
}
