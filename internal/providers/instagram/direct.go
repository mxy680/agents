package instagram

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// directInboxResponse is the response for GET /api/v1/direct_v2/inbox/.
type directInboxResponse struct {
	Inbox  directInbox `json:"inbox"`
	Status string      `json:"status"`
}

// directInbox is the inner inbox object.
type directInbox struct {
	Threads     []rawDirectThread `json:"threads"`
	OldestCursor string           `json:"oldest_cursor"`
	HasOlder    bool              `json:"has_older"`
}

// rawDirectThread is a raw DM thread from the Instagram API.
type rawDirectThread struct {
	ThreadID       string          `json:"thread_id"`
	ThreadTitle    string          `json:"thread_title"`
	LastActivityAt int64           `json:"last_activity_at"`
	IsGroup        bool            `json:"is_group"`
	Users          []rawUser       `json:"users"`
	Items          []rawDirectItem `json:"items"`
}

// rawDirectItem is a single message item in a thread.
type rawDirectItem struct {
	ItemID    string `json:"item_id"`
	ItemType  string `json:"item_type"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	UserID    string `json:"user_id"`
}

// directThreadResponse is the response for GET /api/v1/direct_v2/threads/{thread_id}/.
type directThreadResponse struct {
	Thread directThreadDetail `json:"thread"`
	Status string             `json:"status"`
}

// directThreadDetail is a full thread with its message items.
type directThreadDetail struct {
	ThreadID       string          `json:"thread_id"`
	ThreadTitle    string          `json:"thread_title"`
	LastActivityAt int64           `json:"last_activity_at"`
	IsGroup        bool            `json:"is_group"`
	Users          []rawUser       `json:"users"`
	Items          []rawDirectItem `json:"items"`
	OldestCursor   string          `json:"oldest_cursor"`
	HasOlder       bool            `json:"has_older"`
}

// directActionResponse is a generic response for direct message actions.
type directActionResponse struct {
	Status string `json:"status"`
}

// DirectMessageSummary is the output shape for a DM message item.
type DirectMessageSummary struct {
	ItemID    string `json:"item_id"`
	ItemType  string `json:"item_type"`
	Text      string `json:"text,omitempty"`
	Timestamp int64  `json:"timestamp"`
	UserID    string `json:"user_id"`
}

// newDirectCmd builds the `direct` subcommand group.
func newDirectCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "direct",
		Short:   "Manage direct messages",
		Aliases: []string{"dm", "msg"},
	}
	cmd.AddCommand(newDirectThreadsCmd(factory))
	cmd.AddCommand(newDirectGetCmd(factory))
	cmd.AddCommand(newDirectSendCmd(factory))
	cmd.AddCommand(newDirectCreateCmd(factory))
	cmd.AddCommand(newDirectDeleteMessageCmd(factory))
	cmd.AddCommand(newDirectMarkSeenCmd(factory))
	cmd.AddCommand(newDirectPendingCmd(factory))
	cmd.AddCommand(newDirectApproveCmd(factory))
	cmd.AddCommand(newDirectDeclineCmd(factory))
	return cmd
}

func newDirectThreadsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "threads",
		Short: "List DM threads (inbox)",
		RunE:  makeRunDirectThreads(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of threads to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunDirectThreads(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/direct_v2/inbox/", params)
		if err != nil {
			return fmt.Errorf("listing DM threads: %w", err)
		}

		var result directInboxResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			if strings.Contains(err.Error(), "Prompt has contribution") || strings.Contains(err.Error(), "encryption") {
				return fmt.Errorf("direct messages require end-to-end encryption setup — this is a known limitation of the private API")
			}
			return fmt.Errorf("decoding inbox response: %w", err)
		}

		threads := result.Inbox.Threads
		summaries := make([]DirectThreadSummary, 0, len(threads))
		for _, t := range threads {
			summaries = append(summaries, DirectThreadSummary{
				ThreadID:     t.ThreadID,
				ThreadTitle:  t.ThreadTitle,
				LastActivity: t.LastActivityAt,
				IsGroup:      t.IsGroup,
			})
		}

		if err := printDirectThreadSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.Inbox.HasOlder && result.Inbox.OldestCursor != "" {
			fmt.Printf("Next cursor: %s\n", result.Inbox.OldestCursor)
		}
		return nil
	}
}

func newDirectGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get messages in a DM thread",
		RunE:  makeRunDirectGet(factory),
	}
	cmd.Flags().String("thread-id", "", "Thread ID")
	_ = cmd.MarkFlagRequired("thread-id")
	cmd.Flags().Int("limit", 20, "Maximum number of messages to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunDirectGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		threadID, _ := cmd.Flags().GetString("thread-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/direct_v2/threads/"+threadID+"/", params)
		if err != nil {
			return fmt.Errorf("getting thread %s: %w", threadID, err)
		}

		var result directThreadResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding thread response: %w", err)
		}

		summaries := make([]DirectMessageSummary, 0, len(result.Thread.Items))
		for _, item := range result.Thread.Items {
			summaries = append(summaries, DirectMessageSummary{
				ItemID:    item.ItemID,
				ItemType:  item.ItemType,
				Text:      item.Text,
				Timestamp: item.Timestamp,
				UserID:    item.UserID,
			})
		}

		if err := printDirectMessages(cmd, summaries); err != nil {
			return err
		}
		if result.Thread.HasOlder && result.Thread.OldestCursor != "" {
			fmt.Printf("Next cursor: %s\n", result.Thread.OldestCursor)
		}
		return nil
	}
}

func newDirectSendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a text message to a thread",
		RunE:  makeRunDirectSend(factory),
	}
	cmd.Flags().String("thread-id", "", "Thread ID to send to")
	_ = cmd.MarkFlagRequired("thread-id")
	cmd.Flags().String("text", "", "Message text to send")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunDirectSend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		threadID, _ := cmd.Flags().GetString("thread-id")
		text, _ := cmd.Flags().GetString("text")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("send message to thread %s: %q", threadID, text),
				map[string]string{"thread_id": threadID, "text": text})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("thread_ids", "[\""+threadID+"\"]")
		body.Set("text", text)

		resp, err := client.MobilePost(ctx, "/api/v1/direct_v2/threads/broadcast/text/", body)
		if err != nil {
			return fmt.Errorf("sending message to thread %s: %w", threadID, err)
		}

		var result directActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding send response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message sent to thread %s\n", threadID)
		return nil
	}
}

func newDirectCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new group DM thread",
		RunE:  makeRunDirectCreate(factory),
	}
	cmd.Flags().String("user-ids", "", "Comma-separated user IDs to include in the thread")
	_ = cmd.MarkFlagRequired("user-ids")
	cmd.Flags().String("message", "", "Optional initial message text")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunDirectCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userIDs, _ := cmd.Flags().GetString("user-ids")
		message, _ := cmd.Flags().GetString("message")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("create group thread with users %s", userIDs),
				map[string]string{"user_ids": userIDs, "message": message})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("recipient_users", "[[\""+userIDs+"\"]]")
		if message != "" {
			body.Set("text", message)
		}

		resp, err := client.MobilePost(ctx, "/api/v1/direct_v2/create_group_thread/", body)
		if err != nil {
			return fmt.Errorf("creating group thread: %w", err)
		}

		var result directActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding create thread response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Println("Group thread created")
		return nil
	}
}

func newDirectDeleteMessageCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-message",
		Short: "Unsend (delete) a message from a thread",
		RunE:  makeRunDirectDeleteMessage(factory),
	}
	cmd.Flags().String("thread-id", "", "Thread ID")
	_ = cmd.MarkFlagRequired("thread-id")
	cmd.Flags().String("item-id", "", "Message item ID to delete")
	_ = cmd.MarkFlagRequired("item-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunDirectDeleteMessage(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		threadID, _ := cmd.Flags().GetString("thread-id")
		itemID, _ := cmd.Flags().GetString("item-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("delete message %s from thread %s", itemID, threadID),
				map[string]string{"thread_id": threadID, "item_id": itemID})
		}
		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/direct_v2/threads/"+threadID+"/items/"+itemID+"/delete/", nil)
		if err != nil {
			return fmt.Errorf("deleting message %s from thread %s: %w", itemID, threadID, err)
		}

		var result directActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding delete message response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Deleted message %s from thread %s\n", itemID, threadID)
		return nil
	}
}

func newDirectMarkSeenCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-seen",
		Short: "Mark a thread message as seen",
		RunE:  makeRunDirectMarkSeen(factory),
	}
	cmd.Flags().String("thread-id", "", "Thread ID")
	_ = cmd.MarkFlagRequired("thread-id")
	cmd.Flags().String("item-id", "", "Message item ID to mark as seen")
	_ = cmd.MarkFlagRequired("item-id")
	return cmd
}

func makeRunDirectMarkSeen(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		threadID, _ := cmd.Flags().GetString("thread-id")
		itemID, _ := cmd.Flags().GetString("item-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/direct_v2/threads/"+threadID+"/items/"+itemID+"/seen/", nil)
		if err != nil {
			return fmt.Errorf("marking message %s as seen in thread %s: %w", itemID, threadID, err)
		}

		var result directActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding mark-seen response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Marked message %s as seen in thread %s\n", itemID, threadID)
		return nil
	}
}

func newDirectPendingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending",
		Short: "List pending DM requests",
		RunE:  makeRunDirectPending(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of pending requests to return")
	return cmd
}

func makeRunDirectPending(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))

		resp, err := client.MobileGet(ctx, "/api/v1/direct_v2/pending_inbox/", params)
		if err != nil {
			return fmt.Errorf("listing pending DM requests: %w", err)
		}

		var result directInboxResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding pending inbox response: %w", err)
		}

		threads := result.Inbox.Threads
		summaries := make([]DirectThreadSummary, 0, len(threads))
		for _, t := range threads {
			summaries = append(summaries, DirectThreadSummary{
				ThreadID:    t.ThreadID,
				ThreadTitle: t.ThreadTitle,
				IsGroup:     t.IsGroup,
			})
		}

		return printDirectThreadSummaries(cmd, summaries)
	}
}

func newDirectApproveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve a pending DM request",
		RunE:  makeRunDirectApprove(factory),
	}
	cmd.Flags().String("thread-id", "", "Thread ID of the pending request to approve")
	_ = cmd.MarkFlagRequired("thread-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunDirectApprove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		threadID, _ := cmd.Flags().GetString("thread-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("approve DM request for thread %s", threadID),
				map[string]string{"thread_id": threadID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/direct_v2/threads/"+threadID+"/approve/", nil)
		if err != nil {
			return fmt.Errorf("approving DM request for thread %s: %w", threadID, err)
		}

		var result directActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding approve response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Approved DM request for thread %s\n", threadID)
		return nil
	}
}

func newDirectDeclineCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decline",
		Short: "Decline (hide) a pending DM request",
		RunE:  makeRunDirectDecline(factory),
	}
	cmd.Flags().String("thread-id", "", "Thread ID of the pending request to decline")
	_ = cmd.MarkFlagRequired("thread-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunDirectDecline(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		threadID, _ := cmd.Flags().GetString("thread-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("decline DM request for thread %s", threadID),
				map[string]string{"thread_id": threadID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/direct_v2/threads/"+threadID+"/hide/", nil)
		if err != nil {
			return fmt.Errorf("declining DM request for thread %s: %w", threadID, err)
		}

		var result directActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding decline response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Declined DM request for thread %s\n", threadID)
		return nil
	}
}

// printDirectThreadSummaries outputs DM thread summaries as JSON or text.
func printDirectThreadSummaries(cmd *cobra.Command, threads []DirectThreadSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(threads)
	}

	if len(threads) == 0 {
		fmt.Println("No threads found.")
		return nil
	}

	lines := make([]string, 0, len(threads)+1)
	lines = append(lines, fmt.Sprintf("%-24s  %-30s  %-8s", "THREAD ID", "TITLE", "GROUP"))
	for _, t := range threads {
		lines = append(lines, fmt.Sprintf("%-24s  %-30s  %-8v",
			truncate(t.ThreadID, 24),
			truncate(t.ThreadTitle, 30),
			t.IsGroup,
		))
	}
	cli.PrintText(lines)
	return nil
}

// printDirectMessages outputs DM message summaries as JSON or text.
func printDirectMessages(cmd *cobra.Command, messages []DirectMessageSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(messages)
	}

	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	lines := make([]string, 0, len(messages)+1)
	lines = append(lines, fmt.Sprintf("%-24s  %-10s  %-16s  %-40s", "ITEM ID", "TYPE", "TIMESTAMP", "TEXT"))
	for _, m := range messages {
		lines = append(lines, fmt.Sprintf("%-24s  %-10s  %-16s  %-40s",
			truncate(m.ItemID, 24),
			truncate(m.ItemType, 10),
			formatTimestamp(m.Timestamp),
			truncate(m.Text, 40),
		))
	}
	cli.PrintText(lines)
	return nil
}
