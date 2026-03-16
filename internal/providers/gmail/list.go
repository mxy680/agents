package gmail

import (
	"context"
	"fmt"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// EmailSummary is the JSON-serializable summary of an email.
type EmailSummary struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
}

func newListUnreadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-unread",
		Short: "List unread emails from inbox",
		RunE:  runListUnread,
	}
	cmd.Flags().Int("limit", 20, "Maximum number of messages to return")
	cmd.Flags().String("since", "24h", "Only messages newer than this duration (e.g. 24h, 7d)")
	return cmd
}

// parseSinceDuration parses a duration string like "24h" or "7d".
func parseSinceDuration(s string) (time.Duration, error) {
	// Handle day shorthand
	if len(s) > 0 && s[len(s)-1] == 'd' {
		s = s[:len(s)-1] + "h"
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return d * 24, nil
	}
	return time.ParseDuration(s)
}

func runListUnread(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	limit, _ := cmd.Flags().GetInt("limit")
	since, _ := cmd.Flags().GetString("since")

	dur, err := parseSinceDuration(since)
	if err != nil {
		return fmt.Errorf("invalid --since value: %w", err)
	}

	svc, err := auth.NewGmailService(ctx)
	if err != nil {
		return err
	}

	afterEpoch := time.Now().Add(-dur).Unix()
	query := fmt.Sprintf("is:unread after:%d", afterEpoch)

	resp, err := svc.Users.Messages.List("me").Q(query).MaxResults(int64(limit)).Do()
	if err != nil {
		return fmt.Errorf("listing messages: %w", err)
	}

	summaries, err := fetchSummaries(ctx, svc, resp.Messages)
	if err != nil {
		return err
	}

	return printSummaries(cmd, summaries)
}

func fetchSummaries(_ context.Context, svc *api.Service, msgs []*api.Message) ([]EmailSummary, error) {
	summaries := make([]EmailSummary, 0, len(msgs))
	for _, m := range msgs {
		msg, err := svc.Users.Messages.Get("me", m.Id).Format("metadata").MetadataHeaders("From", "Subject", "Date").Do()
		if err != nil {
			return nil, fmt.Errorf("getting message %s: %w", m.Id, err)
		}
		summary := EmailSummary{
			ID:      msg.Id,
			Snippet: msg.Snippet,
		}
		for _, h := range msg.Payload.Headers {
			switch h.Name {
			case "From":
				summary.From = h.Value
			case "Subject":
				summary.Subject = h.Value
			case "Date":
				summary.Date = h.Value
			}
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func printSummaries(cmd *cobra.Command, summaries []EmailSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No unread messages found.")
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
