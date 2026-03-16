package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// DraftSummary is the JSON-serializable summary of a Gmail draft.
type DraftSummary struct {
	ID        string `json:"id"`
	MessageID string `json:"messageId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Snippet   string `json:"snippet"`
	Date      string `json:"date"`
}

// DraftDetail is the JSON-serializable full content of a Gmail draft.
type DraftDetail struct {
	ID        string `json:"id"`
	MessageID string `json:"messageId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Date      string `json:"date"`
	Body      string `json:"body"`
}

// draftSummaryFromAPI converts a Gmail API Draft to DraftSummary.
func draftSummaryFromAPI(d *api.Draft) DraftSummary {
	summary := DraftSummary{ID: d.Id}
	if d.Message == nil {
		return summary
	}
	summary.MessageID = d.Message.Id
	summary.Snippet = d.Message.Snippet
	if d.Message.Payload != nil {
		headers := extractHeaders(d.Message.Payload.Headers, "From", "To", "Subject", "Date")
		summary.From = headers["From"]
		summary.To = headers["To"]
		summary.Subject = headers["Subject"]
		summary.Date = headers["Date"]
	}
	return summary
}

// newDraftsListCmd returns the `drafts list` command.
func newDraftsListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List drafts in the mailbox",
		RunE:  makeRunDraftsList(factory),
	}
	cmd.Flags().Int64("limit", 20, "Maximum number of drafts to return")
	cmd.Flags().String("page-token", "", "Page token for pagination")
	return cmd
}

func makeRunDraftsList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		limit, _ := cmd.Flags().GetInt64("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := svc.Users.Drafts.List("me").MaxResults(limit)
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing drafts: %w", err)
		}

		summaries := make([]DraftSummary, 0, len(resp.Drafts))
		for _, d := range resp.Drafts {
			summaries = append(summaries, draftSummaryFromAPI(d))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No drafts found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-40s  %s", "ID", "TO", "SUBJECT", "DATE"))
		for _, s := range summaries {
			id := truncate(s.ID, 20)
			to := truncate(s.To, 30)
			subject := truncate(s.Subject, 40)
			lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-40s  %s", id, to, subject, s.Date))
		}
		cli.PrintText(lines)
		return nil
	}
}

// newDraftsGetCmd returns the `drafts get` command.
func newDraftsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a draft by ID",
		RunE:  makeRunDraftsGet(factory),
	}
	cmd.Flags().String("id", "", "Draft ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunDraftsGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		d, err := svc.Users.Drafts.Get("me", id).Do()
		if err != nil {
			return fmt.Errorf("getting draft %s: %w", id, err)
		}

		detail := DraftDetail{ID: d.Id}
		if d.Message != nil {
			detail.MessageID = d.Message.Id
			if d.Message.Payload != nil {
				headers := extractHeaders(d.Message.Payload.Headers, "From", "To", "Subject", "Date")
				detail.From = headers["From"]
				detail.To = headers["To"]
				detail.Subject = headers["Subject"]
				detail.Date = headers["Date"]
				detail.Body = extractBody(d.Message.Payload)
			}
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:        %s", detail.ID),
			fmt.Sprintf("MessageID: %s", detail.MessageID),
			fmt.Sprintf("From:      %s", detail.From),
			fmt.Sprintf("To:        %s", detail.To),
			fmt.Sprintf("Subject:   %s", detail.Subject),
			fmt.Sprintf("Date:      %s", detail.Date),
			"",
			detail.Body,
		}
		cli.PrintText(lines)
		return nil
	}
}

// newDraftsCreateCmd returns the `drafts create` command.
func newDraftsCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new draft",
		RunE:  makeRunDraftsCreate(factory),
	}
	cmd.Flags().String("to", "", "Recipient email address (required)")
	cmd.Flags().String("subject", "", "Email subject (required)")
	cmd.Flags().String("body", "", "Draft body text")
	cmd.Flags().String("body-file", "", "Path to file containing draft body")
	cmd.Flags().String("cc", "", "CC recipient email address")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("subject")
	return cmd
}

func makeRunDraftsCreate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")
		cc, _ := cmd.Flags().GetString("cc")

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

		raw, err := composeMessage(ctx, factory, to, subject, body, cc, "")
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create draft to %s: %s", to, subject), map[string]string{
				"id":        "",
				"messageId": "",
				"status":    "created",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		encoded := base64.URLEncoding.EncodeToString([]byte(raw))
		created, err := svc.Users.Drafts.Create("me", &api.Draft{
			Message: &api.Message{Raw: encoded},
		}).Do()
		if err != nil {
			return fmt.Errorf("creating draft: %w", err)
		}

		msgID := ""
		if created.Message != nil {
			msgID = created.Message.Id
		}
		result := map[string]string{"id": created.Id, "messageId": msgID, "status": "created"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Draft created (id: %s)\n", created.Id)
		return nil
	}
}

// newDraftsUpdateCmd returns the `drafts update` command.
func newDraftsUpdateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing draft with new content",
		RunE:  makeRunDraftsUpdate(factory),
	}
	cmd.Flags().String("id", "", "Draft ID (required)")
	cmd.Flags().String("to", "", "Recipient email address (required)")
	cmd.Flags().String("subject", "", "Email subject (required)")
	cmd.Flags().String("body", "", "Draft body text")
	cmd.Flags().String("body-file", "", "Path to file containing draft body")
	cmd.Flags().String("cc", "", "CC recipient email address")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("subject")
	return cmd
}

func makeRunDraftsUpdate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")
		cc, _ := cmd.Flags().GetString("cc")

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

		raw, err := composeMessage(ctx, factory, to, subject, body, cc, "")
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update draft %s", id), map[string]string{
				"id":        id,
				"messageId": "",
				"status":    "updated",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		encoded := base64.URLEncoding.EncodeToString([]byte(raw))
		updated, err := svc.Users.Drafts.Update("me", id, &api.Draft{
			Message: &api.Message{Raw: encoded},
		}).Do()
		if err != nil {
			return fmt.Errorf("updating draft %s: %w", id, err)
		}

		msgID := ""
		if updated.Message != nil {
			msgID = updated.Message.Id
		}
		result := map[string]string{"id": updated.Id, "messageId": msgID, "status": "updated"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Draft %s updated\n", id)
		return nil
	}
}

// newDraftsSendCmd returns the `drafts send` command.
func newDraftsSendCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an existing draft",
		RunE:  makeRunDraftsSend(factory),
	}
	cmd.Flags().String("id", "", "Draft ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunDraftsSend(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would send draft %s", id), map[string]string{
				"id":       id,
				"threadId": "",
				"status":   "sent",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		sent, err := svc.Users.Drafts.Send("me", &api.Draft{Id: id}).Do()
		if err != nil {
			return fmt.Errorf("sending draft %s: %w", id, err)
		}

		result := map[string]string{"id": sent.Id, "threadId": sent.ThreadId, "status": "sent"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Draft %s sent\n", id)
		return nil
	}
}

// newDraftsDeleteCmd returns the `drafts delete` command.
func newDraftsDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete a draft (IRREVERSIBLE)",
		RunE:  makeRunDraftsDelete(factory),
	}
	cmd.Flags().String("id", "", "Draft ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunDraftsDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would permanently delete draft "+id, map[string]string{"id": id, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Drafts.Delete("me", id).Do()
		if err != nil {
			return fmt.Errorf("deleting draft %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Draft %s permanently deleted\n", id)
		return nil
	}
}
