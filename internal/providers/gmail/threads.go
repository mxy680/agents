package gmail

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// ThreadSummary is the JSON-serializable summary of a thread.
type ThreadSummary struct {
	ID           string `json:"id"`
	Snippet      string `json:"snippet"`
	HistoryID    uint64 `json:"historyId"`
	MessageCount int    `json:"messageCount"`
}

// ThreadDetail is the JSON-serializable detail of a thread with its messages.
type ThreadDetail struct {
	ID        string         `json:"id"`
	HistoryID uint64         `json:"historyId"`
	Messages  []EmailSummary `json:"messages"`
}

// newThreadsListCmd returns the `threads list` command.
func newThreadsListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Gmail threads",
		RunE:  makeRunThreadsList(factory),
	}
	cmd.Flags().String("query", "", "Gmail search query (e.g. is:unread, from:boss)")
	cmd.Flags().Int("limit", 20, "Maximum number of threads to return")
	cmd.Flags().String("page-token", "", "Page token for pagination")
	return cmd
}

func makeRunThreadsList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		req := svc.Users.Threads.List("me").MaxResults(int64(limit))
		if query != "" {
			req = req.Q(query)
		}
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing threads: %w", err)
		}

		summaries := make([]ThreadSummary, 0, len(resp.Threads))
		for _, t := range resp.Threads {
			summaries = append(summaries, ThreadSummary{
				ID:           t.Id,
				Snippet:      t.Snippet,
				HistoryID:    t.HistoryId,
				MessageCount: len(t.Messages),
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No threads found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-60s", "THREAD ID", "SNIPPET"))
		for _, s := range summaries {
			lines = append(lines, fmt.Sprintf("%-20s  %s", s.ID, truncate(s.Snippet, 60)))
		}
		cli.PrintText(lines)
		return nil
	}
}

// newThreadsGetCmd returns the `threads get` command.
func newThreadsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a thread with all its messages",
		RunE:  makeRunThreadsGet(factory),
	}
	cmd.Flags().String("id", "", "Thread ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunThreadsGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		threadID, _ := cmd.Flags().GetString("id")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		thread, err := svc.Users.Threads.Get("me", threadID).Do()
		if err != nil {
			return fmt.Errorf("getting thread %s: %w", threadID, err)
		}

		messages := make([]EmailSummary, 0, len(thread.Messages))
		for _, m := range thread.Messages {
			var headers map[string]string
			if m.Payload != nil {
				headers = extractHeaders(m.Payload.Headers, "From", "Subject", "Date")
			}
			messages = append(messages, EmailSummary{
				ID:      m.Id,
				Snippet: m.Snippet,
				From:    headers["From"],
				Subject: headers["Subject"],
				Date:    headers["Date"],
			})
		}

		detail := ThreadDetail{
			ID:        thread.Id,
			HistoryID: thread.HistoryId,
			Messages:  messages,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Thread: %s", detail.ID),
			fmt.Sprintf("Messages: %d", len(detail.Messages)),
			"",
		}
		for i, m := range detail.Messages {
			lines = append(lines,
				fmt.Sprintf("--- Message %d ---", i+1),
				fmt.Sprintf("ID:      %s", m.ID),
				fmt.Sprintf("From:    %s", m.From),
				fmt.Sprintf("Subject: %s", m.Subject),
				fmt.Sprintf("Date:    %s", m.Date),
				fmt.Sprintf("Snippet: %s", truncate(m.Snippet, 80)),
				"",
			)
		}
		cli.PrintText(lines)
		return nil
	}
}

// newThreadsTrashCmd returns the `threads trash` command.
func newThreadsTrashCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trash",
		Short: "Move a thread to trash",
		RunE:  makeRunThreadsTrash(factory),
	}
	cmd.Flags().String("id", "", "Thread ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunThreadsTrash(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would trash thread "+id, map[string]string{"id": id, "status": "trashed"})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = svc.Users.Threads.Trash("me", id).Do()
		if err != nil {
			return fmt.Errorf("trashing thread %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "trashed"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Thread %s moved to trash\n", id)
		return nil
	}
}

// newThreadsUntrashCmd returns the `threads untrash` command.
func newThreadsUntrashCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untrash",
		Short: "Remove a thread from trash",
		RunE:  makeRunThreadsUntrash(factory),
	}
	cmd.Flags().String("id", "", "Thread ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunThreadsUntrash(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would untrash thread "+id, map[string]string{"id": id, "status": "untrashed"})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = svc.Users.Threads.Untrash("me", id).Do()
		if err != nil {
			return fmt.Errorf("untrashing thread %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "untrashed"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Thread %s removed from trash\n", id)
		return nil
	}
}

// newThreadsDeleteCmd returns the `threads delete` command.
func newThreadsDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete a thread (IRREVERSIBLE)",
		RunE:  makeRunThreadsDelete(factory),
	}
	cmd.Flags().String("id", "", "Thread ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunThreadsDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would permanently delete thread "+id, map[string]string{"id": id, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Threads.Delete("me", id).Do()
		if err != nil {
			return fmt.Errorf("deleting thread %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Thread %s permanently deleted\n", id)
		return nil
	}
}

// newThreadsModifyCmd returns the `threads modify` command.
func newThreadsModifyCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Add or remove labels on all messages in a thread",
		RunE:  makeRunThreadsModify(factory),
	}
	cmd.Flags().String("id", "", "Thread ID (required)")
	cmd.Flags().StringSlice("add-labels", nil, "Labels to add (comma-separated)")
	cmd.Flags().StringSlice("remove-labels", nil, "Labels to remove (comma-separated)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunThreadsModify(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")
		addLabels, _ := cmd.Flags().GetStringSlice("add-labels")
		removeLabels, _ := cmd.Flags().GetStringSlice("remove-labels")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would modify labels on thread %s (add: %v, remove: %v)", id, addLabels, removeLabels), map[string]any{
				"id":     id,
				"status": "modified",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = svc.Users.Threads.Modify("me", id, &api.ModifyThreadRequest{
			AddLabelIds:    addLabels,
			RemoveLabelIds: removeLabels,
		}).Do()
		if err != nil {
			return fmt.Errorf("modifying thread %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "modified"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Thread %s labels updated\n", id)
		return nil
	}
}
