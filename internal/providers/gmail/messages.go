package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// newMessagesListCmd returns the `messages list` command, which replaces both
// `list-unread` and `search`. An optional --query flag accepts any Gmail search
// expression; --since adds an `after:<epoch>` qualifier automatically.
func newMessagesListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Gmail messages",
		Long:  "List Gmail messages. Use --query for Gmail search syntax (e.g. is:unread, from:boss).",
		RunE:  makeRunMessagesList(factory),
	}
	cmd.Flags().String("query", "", "Gmail search query (e.g. is:unread, from:boss)")
	cmd.Flags().Int("limit", 20, "Maximum number of messages to return")
	cmd.Flags().String("since", "", "Only messages newer than this duration (e.g. 24h, 7d)")
	cmd.Flags().String("page-token", "", "Page token for pagination")
	return cmd
}

func makeRunMessagesList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		since, _ := cmd.Flags().GetString("since")
		pageToken, _ := cmd.Flags().GetString("page-token")

		if since != "" {
			dur, err := parseSinceDuration(since)
			if err != nil {
				return fmt.Errorf("invalid --since value: %w", err)
			}
			afterEpoch := time.Now().Add(-dur).Unix()
			afterClause := fmt.Sprintf("after:%d", afterEpoch)
			if query == "" {
				query = afterClause
			} else {
				query = query + " " + afterClause
			}
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := svc.Users.Messages.List("me").MaxResults(int64(limit))
		if query != "" {
			req = req.Q(query)
		}
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing messages: %w", err)
		}

		summaries, err := fetchSummaries(ctx, svc, resp.Messages)
		if err != nil {
			return err
		}

		return printSummaries(cmd, summaries)
	}
}

// newMessagesGetCmd returns the `messages get` command, which replaces `read`.
func newMessagesGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read a full email by message ID",
		RunE:  makeRunMessagesGet(factory),
	}
	cmd.Flags().String("id", "", "Message ID to read (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunMessagesGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		msgID, _ := cmd.Flags().GetString("id")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		msg, err := svc.Users.Messages.Get("me", msgID).Format("full").Do()
		if err != nil {
			return fmt.Errorf("getting message %s: %w", msgID, err)
		}

		detail := extractDetail(msg)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("From:    %s", detail.From),
			fmt.Sprintf("To:      %s", detail.To),
			fmt.Sprintf("Subject: %s", detail.Subject),
			fmt.Sprintf("Date:    %s", detail.Date),
			"",
			detail.Body,
		}
		cli.PrintText(lines)
		return nil
	}
}

// newMessagesSendCmd returns the `messages send` command, keeping parity with
// the old top-level `send` command.
func newMessagesSendCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email via Gmail",
		RunE:  makeRunMessagesSend(factory),
	}
	cmd.Flags().String("to", "", "Recipient email address (required)")
	cmd.Flags().String("subject", "", "Email subject (required)")
	cmd.Flags().String("body", "", "Email body text")
	cmd.Flags().String("body-file", "", "Path to file containing email body")
	cmd.Flags().String("cc", "", "CC recipient email address")
	cmd.Flags().String("reply-to", "", "Message ID to reply to")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("subject")
	return cmd
}

func makeRunMessagesSend(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")
		cc, _ := cmd.Flags().GetString("cc")
		replyTo, _ := cmd.Flags().GetString("reply-to")

		if body == "" && bodyFile == "" {
			return fmt.Errorf("either --body or --body-file is required")
		}

		if bodyFile != "" {
			data, err := os.ReadFile(bodyFile)
			if err != nil {
				return fmt.Errorf("reading body file: %w", err)
			}
			body = string(data)
		}

		raw, err := composeMessage(ctx, factory, to, subject, body, cc, replyTo)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would send:\n%s", raw), map[string]string{
				"status":  "dry-run",
				"raw":     raw,
				"to":      to,
				"subject": subject,
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		msg := &api.Message{
			Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
		}

		sent, err := svc.Users.Messages.Send("me", msg).Do()
		if err != nil {
			return fmt.Errorf("sending message: %w", err)
		}

		result := SendResult{
			ID:       sent.Id,
			ThreadID: sent.ThreadId,
			Status:   "sent",
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Email sent to %s (id: %s)\n", to, sent.Id)
		return nil
	}
}
