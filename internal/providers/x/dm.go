package x

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for DM operations.
const (
	hashDMMessageDelete   = "BJ6DtxA2llfjnRoRjaiIiw"
	hashDMReactionAdd     = "VyDyV9pC2oZEj6g52hgnhA"
	hashDMReactionRemove  = "bV_Nim3RYHsaJwMkTXJ6ew"
	hashDMAddParticipants = "oBwyQ0_xVbAQ8FAyG0pCRA"
)

// newDMCmd builds the "dm" subcommand group.
func newDMCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dm",
		Short: "Manage direct messages",
	}
	cmd.AddCommand(newDMInboxCmd(factory))
	cmd.AddCommand(newDMConversationCmd(factory))
	cmd.AddCommand(newDMSendCmd(factory))
	cmd.AddCommand(newDMSendGroupCmd(factory))
	cmd.AddCommand(newDMDeleteCmd(factory))
	cmd.AddCommand(newDMReactCmd(factory))
	cmd.AddCommand(newDMUnreactCmd(factory))
	cmd.AddCommand(newDMAddMembersCmd(factory))
	cmd.AddCommand(newDMRenameGroupCmd(factory))
	return cmd
}

// newDMInboxCmd builds the "dm inbox" command.
func newDMInboxCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "inbox",
		Short: "Get DM inbox conversations",
		RunE:  makeRunDMInbox(factory),
	}
}

// newDMConversationCmd builds the "dm conversation" command.
func newDMConversationCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conversation",
		Short: "Get messages from a DM conversation",
		RunE:  makeRunDMConversation(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	_ = cmd.MarkFlagRequired("conversation-id")
	cmd.Flags().Int("limit", 50, "Maximum number of messages")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newDMSendCmd builds the "dm send" command.
func newDMSendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a direct message to a user",
		RunE:  makeRunDMSend(factory),
	}
	cmd.Flags().String("user-id", "", "Recipient user ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().String("text", "", "Message text (required)")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().String("media-id", "", "Media ID to attach")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without sending")
	return cmd
}

// newDMSendGroupCmd builds the "dm send-group" command.
func newDMSendGroupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-group",
		Short: "Send a message to an existing group conversation",
		RunE:  makeRunDMSendGroup(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	_ = cmd.MarkFlagRequired("conversation-id")
	cmd.Flags().String("text", "", "Message text (required)")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().String("media-id", "", "Media ID to attach")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without sending")
	return cmd
}

// newDMDeleteCmd builds the "dm delete" command.
func newDMDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a DM message",
		RunE:  makeRunDMDelete(factory),
	}
	cmd.Flags().String("message-id", "", "Message ID to delete (required)")
	_ = cmd.MarkFlagRequired("message-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without deleting")
	return cmd
}

// newDMReactCmd builds the "dm react" command.
func newDMReactCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "react",
		Short: "Add an emoji reaction to a DM message",
		RunE:  makeRunDMReact(factory),
	}
	cmd.Flags().String("message-id", "", "Message ID (required)")
	_ = cmd.MarkFlagRequired("message-id")
	cmd.Flags().String("emoji", "", "Emoji to react with (required)")
	_ = cmd.MarkFlagRequired("emoji")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without sending")
	return cmd
}

// newDMUnreactCmd builds the "dm unreact" command.
func newDMUnreactCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unreact",
		Short: "Remove an emoji reaction from a DM message",
		RunE:  makeRunDMUnreact(factory),
	}
	cmd.Flags().String("message-id", "", "Message ID (required)")
	_ = cmd.MarkFlagRequired("message-id")
	cmd.Flags().String("emoji", "", "Emoji to remove (required)")
	_ = cmd.MarkFlagRequired("emoji")
	cmd.Flags().Bool("dry-run", false, "Print what would be removed without removing")
	return cmd
}

// newDMAddMembersCmd builds the "dm add-members" command.
func newDMAddMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-members",
		Short: "Add members to a group DM conversation",
		RunE:  makeRunDMAddMembers(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	_ = cmd.MarkFlagRequired("conversation-id")
	cmd.Flags().StringSlice("user-ids", nil, "Comma-separated user IDs to add (required)")
	_ = cmd.MarkFlagRequired("user-ids")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without doing it")
	return cmd
}

// newDMRenameGroupCmd builds the "dm rename-group" command.
func newDMRenameGroupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename-group",
		Short: "Rename a group DM conversation",
		RunE:  makeRunDMRenameGroup(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	_ = cmd.MarkFlagRequired("conversation-id")
	cmd.Flags().String("name", "", "New conversation name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without doing it")
	return cmd
}

// --- RunE implementations ---

func makeRunDMInbox(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("include_profile_interstitial_type", "1")
		params.Set("include_blocking", "1")
		params.Set("include_blocked_by", "1")
		params.Set("include_followed_by", "1")
		params.Set("include_want_retweets", "1")
		params.Set("include_mute_edge", "1")
		params.Set("include_can_dm", "1")
		params.Set("include_can_media_tag", "1")
		params.Set("skip_status", "1")
		params.Set("dm_secret_conversations_enabled", "false")
		params.Set("krs_registration_enabled", "true")
		params.Set("cards_platform", "Web-12")
		params.Set("include_cards", "1")
		params.Set("include_ext_alt_text", "true")
		params.Set("include_quote_count", "true")
		params.Set("include_reply_count", "1")
		params.Set("tweet_mode", "extended")
		params.Set("include_ext_collab_control", "true")
		params.Set("include_groups", "true")
		params.Set("dm_users", "false")
		params.Set("include_inbox_timelines", "true")
		params.Set("include_ext_media_color", "true")
		params.Set("supports_reactions", "true")

		resp, err := client.Get(ctx, "/i/api/1.1/dm/inbox_initial_state.json", params)
		if err != nil {
			return fmt.Errorf("fetching DM inbox: %w", err)
		}

		var raw struct {
			InboxInitialState struct {
				Conversations map[string]json.RawMessage `json:"conversations"`
				Entries       []json.RawMessage          `json:"entries"`
			} `json:"inbox_initial_state"`
		}
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decode DM inbox: %w", err)
		}

		conversations := make([]DMConversationSummary, 0, len(raw.InboxInitialState.Conversations))
		for id, convRaw := range raw.InboxInitialState.Conversations {
			var conv struct {
				Type         string `json:"type"`
				Participants []struct {
					UserID string `json:"user_id"`
				} `json:"participants"`
			}
			if err := json.Unmarshal(convRaw, &conv); err != nil {
				continue
			}
			participants := make([]string, 0, len(conv.Participants))
			for _, p := range conv.Participants {
				participants = append(participants, p.UserID)
			}
			conversations = append(conversations, DMConversationSummary{
				ID:           id,
				Type:         conv.Type,
				Participants: participants,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"conversations": conversations})
		}

		if len(conversations) == 0 {
			fmt.Println("No DM conversations found.")
			return nil
		}
		lines := make([]string, 0, len(conversations)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-15s  %-30s", "ID", "TYPE", "PARTICIPANTS"))
		for _, c := range conversations {
			lines = append(lines, fmt.Sprintf("%-30s  %-15s  %-30s",
				truncate(c.ID, 30),
				truncate(c.Type, 15),
				truncate(strings.Join(c.Participants, ","), 30),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunDMConversation(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		conversationID, _ := cmd.Flags().GetString("conversation-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", fmt.Sprintf("%d", limit))
		params.Set("cards_platform", "Web-12")
		params.Set("include_cards", "1")
		params.Set("include_ext_alt_text", "true")
		params.Set("tweet_mode", "extended")
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		path := fmt.Sprintf("/i/api/1.1/dm/conversation/%s.json", url.PathEscape(conversationID))
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("fetching conversation %s: %w", conversationID, err)
		}

		var raw struct {
			ConversationTimeline struct {
				Entries []struct {
					Message *struct {
						ID             string `json:"id"`
						ConversationID string `json:"conversation_id"`
						MessageData    struct {
							Text     string `json:"text"`
							SenderID string `json:"sender_id"`
							Time     string `json:"time"`
						} `json:"message_data"`
					} `json:"message"`
				} `json:"entries"`
				MinEntryID string `json:"min_entry_id"`
			} `json:"conversation_timeline"`
		}
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decode conversation: %w", err)
		}

		messages := make([]DMMessageSummary, 0, len(raw.ConversationTimeline.Entries))
		for _, entry := range raw.ConversationTimeline.Entries {
			if entry.Message == nil {
				continue
			}
			m := entry.Message
			messages = append(messages, DMMessageSummary{
				ID:             m.ID,
				ConversationID: m.ConversationID,
				SenderID:       m.MessageData.SenderID,
				Text:           m.MessageData.Text,
				Timestamp:      m.MessageData.Time,
			})
		}

		nextCursor := raw.ConversationTimeline.MinEntryID

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"messages":    messages,
				"next_cursor": nextCursor,
			})
		}

		if len(messages) == 0 {
			fmt.Println("No messages found.")
			return nil
		}
		lines := make([]string, 0, len(messages)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-60s", "ID", "SENDER", "TEXT"))
		for _, msg := range messages {
			lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-60s",
				truncate(msg.ID, 20),
				truncate(msg.SenderID, 15),
				truncate(msg.Text, 60),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunDMSend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		text, _ := cmd.Flags().GetString("text")
		mediaID, _ := cmd.Flags().GetString("media-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		body := map[string]any{
			"text":          text,
			"recipient_ids": []string{userID},
		}
		if mediaID != "" {
			body["media_id"] = mediaID
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("send DM to user %s: %q", userID, truncate(text, 60)), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.PostJSON(ctx, "/i/api/1.1/dm/new2.json", body)
		if err != nil {
			return fmt.Errorf("sending DM to %s: %w", userID, err)
		}

		var result map[string]json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decode send DM response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "sent", "recipient_id": userID})
		}
		fmt.Printf("DM sent to user %s\n", userID)
		return nil
	}
}

func makeRunDMSendGroup(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		conversationID, _ := cmd.Flags().GetString("conversation-id")
		text, _ := cmd.Flags().GetString("text")
		mediaID, _ := cmd.Flags().GetString("media-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		body := map[string]any{
			"text":            text,
			"conversation_id": conversationID,
		}
		if mediaID != "" {
			body["media_id"] = mediaID
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("send group DM to conversation %s: %q", conversationID, truncate(text, 60)), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.PostJSON(ctx, "/i/api/1.1/dm/new2.json", body)
		if err != nil {
			return fmt.Errorf("sending group DM to conversation %s: %w", conversationID, err)
		}

		var result map[string]json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decode send group DM response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "sent", "conversation_id": conversationID})
		}
		fmt.Printf("Group DM sent to conversation %s\n", conversationID)
		return nil
	}
}

func makeRunDMDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		messageID, _ := cmd.Flags().GetString("message-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("delete DM message %s", messageID), map[string]string{"message_id": messageID})
		}

		if err := confirmDestructive(cmd, "this action is irreversible"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"messageId": messageID,
		}

		_, err = client.GraphQLPost(ctx, hashDMMessageDelete, "DMMessageDeleteMutation", vars, map[string]bool{})
		if err != nil {
			return fmt.Errorf("deleting DM message %s: %w", messageID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "message_id": messageID})
		}
		fmt.Printf("DM message deleted: %s\n", messageID)
		return nil
	}
}

func makeRunDMReact(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		messageID, _ := cmd.Flags().GetString("message-id")
		emoji, _ := cmd.Flags().GetString("emoji")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"messageId":    messageID,
			"reactionType": emoji,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("react to DM message %s with %s", messageID, emoji), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashDMReactionAdd, "useDMReactionMutationAddMutation", vars, map[string]bool{})
		if err != nil {
			return fmt.Errorf("reacting to DM message %s: %w", messageID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "reacted", "message_id": messageID, "emoji": emoji})
		}
		fmt.Printf("Reacted to DM message %s with %s\n", messageID, emoji)
		return nil
	}
}

func makeRunDMUnreact(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		messageID, _ := cmd.Flags().GetString("message-id")
		emoji, _ := cmd.Flags().GetString("emoji")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"messageId":    messageID,
			"reactionType": emoji,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("remove reaction %s from DM message %s", emoji, messageID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashDMReactionRemove, "useDMReactionMutationRemoveMutation", vars, map[string]bool{})
		if err != nil {
			return fmt.Errorf("removing reaction from DM message %s: %w", messageID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unreacted", "message_id": messageID, "emoji": emoji})
		}
		fmt.Printf("Removed reaction %s from DM message %s\n", emoji, messageID)
		return nil
	}
}

func makeRunDMAddMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		conversationID, _ := cmd.Flags().GetString("conversation-id")
		userIDs, _ := cmd.Flags().GetStringSlice("user-ids")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"conversationId": conversationID,
			"userIds":        userIDs,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("add members %v to conversation %s", userIDs, conversationID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashDMAddParticipants, "AddParticipantsMutation", vars, map[string]bool{})
		if err != nil {
			return fmt.Errorf("adding members to conversation %s: %w", conversationID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"status": "added", "conversation_id": conversationID, "user_ids": userIDs})
		}
		fmt.Printf("Members added to conversation %s\n", conversationID)
		return nil
	}
}

func makeRunDMRenameGroup(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		conversationID, _ := cmd.Flags().GetString("conversation-id")
		name, _ := cmd.Flags().GetString("name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("rename conversation %s to %q", conversationID, name), map[string]string{"conversation_id": conversationID, "name": name})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		formBody := url.Values{}
		formBody.Set("name", name)

		path := fmt.Sprintf("/i/api/1.1/dm/conversation/%s/update_name.json", url.PathEscape(conversationID))
		resp, err := client.Post(ctx, path, formBody)
		if err != nil {
			return fmt.Errorf("renaming conversation %s: %w", conversationID, err)
		}

		var result map[string]json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decode rename response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "renamed", "conversation_id": conversationID, "name": name})
		}
		fmt.Printf("Conversation %s renamed to %q\n", conversationID, name)
		return nil
	}
}
