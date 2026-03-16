package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
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

// newMessagesTrashCmd returns the `messages trash` command.
func newMessagesTrashCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trash",
		Short: "Move a message to trash",
		RunE:  makeRunMessagesTrash(factory),
	}
	cmd.Flags().String("id", "", "Message ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunMessagesTrash(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would trash message "+id, map[string]string{"id": id, "status": "trashed"})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = svc.Users.Messages.Trash("me", id).Do()
		if err != nil {
			return fmt.Errorf("trashing message %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "trashed"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message %s moved to trash\n", id)
		return nil
	}
}

// newMessagesUntrashCmd returns the `messages untrash` command.
func newMessagesUntrashCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untrash",
		Short: "Remove a message from trash",
		RunE:  makeRunMessagesUntrash(factory),
	}
	cmd.Flags().String("id", "", "Message ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunMessagesUntrash(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would untrash message "+id, map[string]string{"id": id, "status": "untrashed"})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = svc.Users.Messages.Untrash("me", id).Do()
		if err != nil {
			return fmt.Errorf("untrashing message %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "untrashed"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message %s removed from trash\n", id)
		return nil
	}
}

// newMessagesDeleteCmd returns the `messages delete` command.
func newMessagesDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete a message (IRREVERSIBLE)",
		RunE:  makeRunMessagesDelete(factory),
	}
	cmd.Flags().String("id", "", "Message ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunMessagesDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would permanently delete message "+id, map[string]string{"id": id, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Messages.Delete("me", id).Do()
		if err != nil {
			return fmt.Errorf("deleting message %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message %s permanently deleted\n", id)
		return nil
	}
}

// newMessagesModifyCmd returns the `messages modify` command.
func newMessagesModifyCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Add or remove labels on a message",
		RunE:  makeRunMessagesModify(factory),
	}
	cmd.Flags().String("id", "", "Message ID (required)")
	cmd.Flags().StringSlice("add-labels", nil, "Labels to add (comma-separated)")
	cmd.Flags().StringSlice("remove-labels", nil, "Labels to remove (comma-separated)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunMessagesModify(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")
		addLabels, _ := cmd.Flags().GetStringSlice("add-labels")
		removeLabels, _ := cmd.Flags().GetStringSlice("remove-labels")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would modify labels on message %s (add: %v, remove: %v)", id, addLabels, removeLabels), map[string]any{
				"id":     id,
				"status": "modified",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		msg, err := svc.Users.Messages.Modify("me", id, &api.ModifyMessageRequest{
			AddLabelIds:    addLabels,
			RemoveLabelIds: removeLabels,
		}).Do()
		if err != nil {
			return fmt.Errorf("modifying message %s: %w", id, err)
		}

		result := map[string]any{"id": msg.Id, "labelIds": msg.LabelIds, "status": "modified"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message %s labels updated\n", id)
		return nil
	}
}

// newMessagesImportCmd returns the `messages import` command.
func newMessagesImportCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a raw RFC 2822 message into the mailbox",
		RunE:  makeRunMessagesImport(factory),
	}
	cmd.Flags().String("raw-file", "", "Path to raw RFC 2822 message file (required)")
	_ = cmd.MarkFlagRequired("raw-file")
	return cmd
}

func makeRunMessagesImport(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		rawFile, _ := cmd.Flags().GetString("raw-file")

		data, err := os.ReadFile(rawFile)
		if err != nil {
			return fmt.Errorf("reading raw file: %w", err)
		}
		encoded := base64.URLEncoding.EncodeToString(data)

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would import message from "+rawFile, map[string]string{"status": "imported"})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		imported, err := svc.Users.Messages.Import("me", &api.Message{Raw: encoded}).Do()
		if err != nil {
			return fmt.Errorf("importing message: %w", err)
		}

		result := map[string]string{"id": imported.Id, "threadId": imported.ThreadId, "status": "imported"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message imported (id: %s)\n", imported.Id)
		return nil
	}
}

// newMessagesInsertCmd returns the `messages insert` command.
func newMessagesInsertCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insert",
		Short: "Insert a raw RFC 2822 message directly into the mailbox without sending",
		RunE:  makeRunMessagesInsert(factory),
	}
	cmd.Flags().String("raw-file", "", "Path to raw RFC 2822 message file (required)")
	_ = cmd.MarkFlagRequired("raw-file")
	return cmd
}

func makeRunMessagesInsert(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		rawFile, _ := cmd.Flags().GetString("raw-file")

		data, err := os.ReadFile(rawFile)
		if err != nil {
			return fmt.Errorf("reading raw file: %w", err)
		}
		encoded := base64.URLEncoding.EncodeToString(data)

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would insert message from "+rawFile, map[string]string{"status": "inserted"})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		inserted, err := svc.Users.Messages.Insert("me", &api.Message{Raw: encoded}).Do()
		if err != nil {
			return fmt.Errorf("inserting message: %w", err)
		}

		result := map[string]string{"id": inserted.Id, "threadId": inserted.ThreadId, "status": "inserted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message inserted (id: %s)\n", inserted.Id)
		return nil
	}
}

// newMessagesBatchModifyCmd returns the `messages batch-modify` command.
func newMessagesBatchModifyCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-modify",
		Short: "Add or remove labels on multiple messages",
		RunE:  makeRunMessagesBatchModify(factory),
	}
	cmd.Flags().String("ids", "", "Comma-separated message IDs (required)")
	cmd.Flags().StringSlice("add-labels", nil, "Labels to add (comma-separated)")
	cmd.Flags().StringSlice("remove-labels", nil, "Labels to remove (comma-separated)")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}

func makeRunMessagesBatchModify(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		idsStr, _ := cmd.Flags().GetString("ids")
		ids := strings.Split(idsStr, ",")
		addLabels, _ := cmd.Flags().GetStringSlice("add-labels")
		removeLabels, _ := cmd.Flags().GetStringSlice("remove-labels")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would batch-modify %d messages (add: %v, remove: %v)", len(ids), addLabels, removeLabels), map[string]any{
				"ids":    ids,
				"status": "modified",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Messages.BatchModify("me", &api.BatchModifyMessagesRequest{
			Ids:            ids,
			AddLabelIds:    addLabels,
			RemoveLabelIds: removeLabels,
		}).Do()
		if err != nil {
			return fmt.Errorf("batch-modifying messages: %w", err)
		}

		result := map[string]any{"ids": ids, "status": "modified"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("%d messages labels updated\n", len(ids))
		return nil
	}
}

// newMessagesBatchDeleteCmd returns the `messages batch-delete` command.
func newMessagesBatchDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-delete",
		Short: "Permanently delete multiple messages (IRREVERSIBLE)",
		RunE:  makeRunMessagesBatchDelete(factory),
	}
	cmd.Flags().String("ids", "", "Comma-separated message IDs (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("ids")
	return cmd
}

func makeRunMessagesBatchDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		idsStr, _ := cmd.Flags().GetString("ids")
		ids := strings.Split(idsStr, ",")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete %d messages", len(ids)), map[string]any{
				"ids":    ids,
				"status": "deleted",
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Messages.BatchDelete("me", &api.BatchDeleteMessagesRequest{
			Ids: ids,
		}).Do()
		if err != nil {
			return fmt.Errorf("batch-deleting messages: %w", err)
		}

		result := map[string]any{"ids": ids, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("%d messages permanently deleted\n", len(ids))
		return nil
	}
}
