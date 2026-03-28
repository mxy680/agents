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
	cmd.AddCommand(newConversationsUnreadCountCmd(factory))
	cmd.AddCommand(newConversationsCreateCmd(factory))
	cmd.AddCommand(newConversationsReplyCmd(factory))
	cmd.AddCommand(newConversationsUpdateCmd(factory))
	cmd.AddCommand(newConversationsDeleteCmd(factory))
	cmd.AddCommand(newConversationsMarkAllReadCmd(factory))

	return cmd
}

func newConversationsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new conversation (send a message)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			recipients, _ := cmd.Flags().GetString("recipients")
			subject, _ := cmd.Flags().GetString("subject")
			body, _ := cmd.Flags().GetString("body")
			if recipients == "" {
				return fmt.Errorf("--recipients is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "create conversation: "+subject, map[string]any{
					"recipients": recipients, "subject": subject, "body": body,
				})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			payload := map[string]any{
				"recipients": strings.Split(recipients, ","),
				"subject":    subject,
				"body":       body,
			}
			data, err := client.Post(ctx, "/conversations", payload)
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
			if len(conversations) > 0 {
				fmt.Printf("Conversation %d created: %s\n", conversations[0].ID, conversations[0].Subject)
			}
			return nil
		},
	}

	cmd.Flags().String("recipients", "", "Comma-separated recipient IDs (required)")
	cmd.Flags().String("subject", "", "Conversation subject")
	cmd.Flags().String("body", "", "Message body")
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

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "reply to conversation "+conversationID, map[string]any{
					"conversation_id": conversationID, "body": body,
				})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			payload := map[string]any{"body": body}
			data, err := client.Post(ctx, "/conversations/"+conversationID+"/add_message", payload)
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
			fmt.Printf("Reply added to conversation %s\n", conversationID)
			return nil
		},
	}

	cmd.Flags().String("conversation-id", "", "Canvas conversation ID (required)")
	cmd.Flags().String("body", "", "Reply message body")
	return cmd
}

func newConversationsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update conversation properties (starred, workflow state)",
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

			payload := map[string]any{}
			if starred, _ := cmd.Flags().GetBool("starred"); starred {
				payload["conversation[starred]"] = true
			}
			if workflowState, _ := cmd.Flags().GetString("workflow-state"); workflowState != "" {
				payload["conversation[workflow_state]"] = workflowState
			}

			data, err := client.Put(ctx, "/conversations/"+conversationID, payload)
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
			fmt.Printf("Conversation %s updated\n", conversationID)
			return nil
		},
	}

	cmd.Flags().String("conversation-id", "", "Canvas conversation ID (required)")
	cmd.Flags().Bool("starred", false, "Mark as starred")
	cmd.Flags().String("workflow-state", "", "New workflow state (read, unread, archived)")
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

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, "/conversations/"+conversationID); err != nil {
				return err
			}

			fmt.Printf("Conversation %s deleted\n", conversationID)
			return nil
		},
	}

	cmd.Flags().String("conversation-id", "", "Canvas conversation ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	return cmd
}

func newConversationsMarkAllReadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-all-read",
		Short: "Mark all conversations as read",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			if _, err := client.Post(ctx, "/conversations/mark_all_as_read", nil); err != nil {
				return err
			}

			fmt.Println("All conversations marked as read")
			return nil
		},
	}

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
