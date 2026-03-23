package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newConversationsCmd returns the parent "conversations" command with all subcommands attached.
func newConversationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "conversations",
		Short:   "Manage Canvas conversations (inbox messages)",
		Aliases: []string{"conv", "msg"},
	}

	cmd.AddCommand(newConversationsListCmd(factory))
	cmd.AddCommand(newConversationsGetCmd(factory))
	cmd.AddCommand(newConversationsCreateCmd(factory))
	cmd.AddCommand(newConversationsReplyCmd(factory))
	cmd.AddCommand(newConversationsUpdateCmd(factory))
	cmd.AddCommand(newConversationsDeleteCmd(factory))
	cmd.AddCommand(newConversationsMarkAllReadCmd(factory))
	cmd.AddCommand(newConversationsUnreadCountCmd(factory))

	return cmd
}

func newConversationsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List conversations in the inbox",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			scope, _ := cmd.Flags().GetString("scope")
			filter, _ := cmd.Flags().GetString("filter")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if scope != "" {
				params.Set("scope", scope)
			}
			if filter != "" {
				params.Add("filter[]", filter)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/conversations", params)
			if err != nil {
				return err
			}

			var conversations []ConversationSummary
			if err := json.Unmarshal(data, &conversations); err != nil {
				return fmt.Errorf("parse conversations: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(conversations)
			}

			if len(conversations) == 0 {
				fmt.Println("No conversations found.")
				return nil
			}
			for _, c := range conversations {
				starred := ""
				if c.Starred {
					starred = "*"
				}
				fmt.Printf("%-6d  %-6s  %-8s%s  %s\n",
					c.ID, c.WorkflowState, starred, c.LastMessageAt, truncate(c.Subject, 50))
			}
			return nil
		},
	}

	cmd.Flags().String("scope", "", "Filter by scope: unread, starred, archived, or sent")
	cmd.Flags().String("filter", "", "Filter by course ID (e.g. course_123)")
	cmd.Flags().Int("limit", 0, "Maximum number of conversations to return")
	return cmd
}

func newConversationsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific conversation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			conversationID, _ := cmd.Flags().GetString("conversation-id")
			if conversationID == "" {
				return fmt.Errorf("--conversation-id is required")
			}

			data, err := client.Get(ctx, "/conversations/"+conversationID, nil)
			if err != nil {
				return err
			}

			var conversation ConversationSummary
			if err := json.Unmarshal(data, &conversation); err != nil {
				return fmt.Errorf("parse conversation: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(conversation)
			}

			fmt.Printf("ID:           %d\n", conversation.ID)
			fmt.Printf("Subject:      %s\n", conversation.Subject)
			fmt.Printf("State:        %s\n", conversation.WorkflowState)
			fmt.Printf("Messages:     %d\n", conversation.MessageCount)
			if conversation.Starred {
				fmt.Println("Starred:      yes")
			}
			if conversation.LastMessageAt != "" {
				fmt.Printf("Last Message: %s\n", conversation.LastMessageAt)
			}
			if conversation.LastMessage != "" {
				fmt.Printf("Preview:      %s\n", truncate(conversation.LastMessage, 200))
			}
			if len(conversation.Participants) > 0 {
				fmt.Printf("Participants: %s\n", strings.Join(conversation.Participants, ", "))
			}
			return nil
		},
	}

	cmd.Flags().String("conversation-id", "", "Canvas conversation ID (required)")
	return cmd
}

func newConversationsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new conversation (send a message)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			recipients, _ := cmd.Flags().GetStringSlice("recipients")
			subject, _ := cmd.Flags().GetString("subject")
			body, _ := cmd.Flags().GetString("body")
			if len(recipients) == 0 {
				return fmt.Errorf("--recipients is required")
			}
			if subject == "" {
				return fmt.Errorf("--subject is required")
			}
			if body == "" {
				return fmt.Errorf("--body is required")
			}

			groupConversation, _ := cmd.Flags().GetBool("group-conversation")
			contextCode, _ := cmd.Flags().GetString("context-code")

			reqBody := map[string]any{
				"recipients": recipients,
				"subject":    subject,
				"body":       body,
			}
			if groupConversation {
				reqBody["group_conversation"] = true
			}
			if contextCode != "" {
				reqBody["context_code"] = contextCode
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create conversation %q to %s", subject, strings.Join(recipients, ", ")), reqBody)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/conversations", reqBody)
			if err != nil {
				return err
			}

			// Canvas returns an array of created conversations.
			var conversations []ConversationSummary
			if err := json.Unmarshal(data, &conversations); err != nil {
				return fmt.Errorf("parse conversation: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(conversations)
			}

			if len(conversations) > 0 {
				fmt.Printf("Created conversation %d: %s\n", conversations[0].ID, conversations[0].Subject)
			} else {
				fmt.Println("Conversation created.")
			}
			return nil
		},
	}

	cmd.Flags().StringSlice("recipients", nil, "Recipient user IDs (required)")
	cmd.Flags().String("subject", "", "Conversation subject (required)")
	cmd.Flags().String("body", "", "Message body (required)")
	cmd.Flags().Bool("group-conversation", false, "Send as a group conversation")
	cmd.Flags().String("context-code", "", "Course context code, e.g. course_123")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newConversationsReplyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Reply to an existing conversation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			conversationID, _ := cmd.Flags().GetString("conversation-id")
			body, _ := cmd.Flags().GetString("body")
			if conversationID == "" {
				return fmt.Errorf("--conversation-id is required")
			}
			if body == "" {
				return fmt.Errorf("--body is required")
			}

			reqBody := map[string]any{"body": body}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("reply to conversation %s", conversationID), reqBody)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/conversations/"+conversationID+"/add_message", reqBody)
			if err != nil {
				return err
			}

			var conversation ConversationSummary
			if err := json.Unmarshal(data, &conversation); err != nil {
				return fmt.Errorf("parse conversation: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(conversation)
			}

			fmt.Printf("Reply sent to conversation %d.\n", conversation.ID)
			return nil
		},
	}

	cmd.Flags().String("conversation-id", "", "Canvas conversation ID (required)")
	cmd.Flags().String("body", "", "Reply message body (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newConversationsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a conversation (star, archive, change state)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			conversationID, _ := cmd.Flags().GetString("conversation-id")
			if conversationID == "" {
				return fmt.Errorf("--conversation-id is required")
			}

			convBody := map[string]any{}
			if cmd.Flags().Changed("starred") {
				v, _ := cmd.Flags().GetBool("starred")
				convBody["starred"] = v
			}
			if cmd.Flags().Changed("subscribed") {
				v, _ := cmd.Flags().GetBool("subscribed")
				convBody["subscribed"] = v
			}
			if cmd.Flags().Changed("workflow-state") {
				v, _ := cmd.Flags().GetString("workflow-state")
				convBody["workflow_state"] = v
			}

			reqBody := map[string]any{"conversation": convBody}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update conversation %s", conversationID), reqBody)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/conversations/"+conversationID, reqBody)
			if err != nil {
				return err
			}

			var conversation ConversationSummary
			if err := json.Unmarshal(data, &conversation); err != nil {
				return fmt.Errorf("parse conversation: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(conversation)
			}

			fmt.Printf("Updated conversation %d.\n", conversation.ID)
			return nil
		},
	}

	cmd.Flags().String("conversation-id", "", "Canvas conversation ID (required)")
	cmd.Flags().Bool("starred", false, "Star the conversation")
	cmd.Flags().Bool("subscribed", false, "Subscribe to the conversation")
	cmd.Flags().String("workflow-state", "", "New state: read, unread, or archived")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newConversationsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a conversation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			conversationID, _ := cmd.Flags().GetString("conversation-id")
			if conversationID == "" {
				return fmt.Errorf("--conversation-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the conversation"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete conversation %s", conversationID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/conversations/"+conversationID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "conversation_id": conversationID})
			}
			fmt.Printf("Conversation %s deleted.\n", conversationID)
			return nil
		},
	}

	cmd.Flags().String("conversation-id", "", "Canvas conversation ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newConversationsMarkAllReadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-all-read",
		Short: "Mark all conversations as read",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "mark all conversations as read", nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Post(ctx, "/conversations/mark_all_as_read", nil)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"marked_read": true})
			}
			fmt.Println("All conversations marked as read.")
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newConversationsUnreadCountCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unread-count",
		Short: "Get the number of unread conversations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/conversations/unread_count", nil)
			if err != nil {
				return err
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse unread count: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}

			count, _ := result["unread_count"]
			fmt.Printf("Unread conversations: %v\n", count)
			return nil
		},
	}

	return cmd
}
