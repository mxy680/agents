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
