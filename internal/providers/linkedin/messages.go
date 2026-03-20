package linkedin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// messengerConversationsQueryID is the current known queryId for the messenger conversations GraphQL endpoint.
const messengerConversationsQueryID = "messengerConversations.0d5e6781bbee71c3e51c8843c6519f48"

// messengerConversationEntity is the GraphQL normalized entity for a conversation.
type messengerConversationEntity struct {
	EntityURN      string `json:"entityUrn"`
	Title          string `json:"title"`
	LastActivityAt int64  `json:"lastActivityAt"`
	UnreadCount    int    `json:"unreadCount"`
}

// voyagerConversationElement is kept for backward-compatibility with toConversationSummary
// and other commands that still use the legacy messaging API.
type voyagerConversationElement struct {
	EntityURN      string                       `json:"entityUrn"`
	ConversationID string                       `json:"conversationId"`
	LastActivityAt int64                        `json:"lastActivityAt"`
	UnreadCount    int                          `json:"unreadCount"`
	Participants   []voyagerMessagingMemberWrap `json:"participants"`
}

type voyagerMessagingMemberWrap struct {
	MessagingMember voyagerMessagingMember `json:"com.linkedin.voyager.messaging.MessagingMember"`
}

type voyagerMessagingMember struct {
	MiniProfile voyagerMiniProfile `json:"miniProfile"`
}

type voyagerMiniProfile struct {
	EntityURN string `json:"entityUrn"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// voyagerMessagesResponse is the response envelope for listing messages in a conversation.
type voyagerMessagesResponse struct {
	Elements []voyagerMessageElement `json:"elements"`
	Paging   voyagerPaging           `json:"paging"`
}

type voyagerMessageElement struct {
	EntityURN    string                     `json:"entityUrn"`
	From         voyagerMessagingMemberWrap `json:"from"`
	EventContent voyagerMessageEventWrap    `json:"eventContent"`
	CreatedAt    int64                      `json:"createdAt"`
}

type voyagerMessageEventWrap struct {
	MessageEvent voyagerMessageEvent `json:"com.linkedin.voyager.messaging.event.MessageEvent"`
}

type voyagerMessageEvent struct {
	Body string `json:"body"`
}

// voyagerPaging is the common paging envelope used in Voyager list responses.
type voyagerPaging struct {
	Start int `json:"start"`
	Count int `json:"count"`
	Total int `json:"total"`
}

// toConversationSummary maps a raw conversation element to ConversationSummary.
func toConversationSummary(el voyagerConversationElement) ConversationSummary {
	participantURNs := make([]string, 0, len(el.Participants))
	names := make([]string, 0, len(el.Participants))
	for _, p := range el.Participants {
		urn := p.MessagingMember.MiniProfile.EntityURN
		if urn != "" {
			participantURNs = append(participantURNs, urn)
		}
		fn := p.MessagingMember.MiniProfile.FirstName
		ln := p.MessagingMember.MiniProfile.LastName
		full := strings.TrimSpace(fn + " " + ln)
		if full != "" {
			names = append(names, full)
		}
	}
	title := strings.Join(names, ", ")
	return ConversationSummary{
		ID:              el.ConversationID,
		Title:           title,
		LastActivityAt:  el.LastActivityAt,
		UnreadCount:     el.UnreadCount,
		ParticipantURNs: participantURNs,
	}
}

// toMessageSummary maps a raw message element to MessageSummary.
func toMessageSummary(el voyagerMessageElement) MessageSummary {
	mp := el.From.MessagingMember.MiniProfile
	senderName := strings.TrimSpace(mp.FirstName + " " + mp.LastName)
	return MessageSummary{
		ID:         el.EntityURN,
		SenderURN:  mp.EntityURN,
		SenderName: senderName,
		Text:       el.EventContent.MessageEvent.Body,
		Timestamp:  el.CreatedAt,
	}
}

// newMessagesCmd builds the "messages" subcommand group.
func newMessagesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messages",
		Short:   "Manage LinkedIn messages",
		Aliases: []string{"msg"},
	}
	cmd.AddCommand(newMessagesConversationsCmd(factory))
	cmd.AddCommand(newMessagesListCmd(factory))
	cmd.AddCommand(newMessagesSendCmd(factory))
	cmd.AddCommand(newMessagesNewCmd(factory))
	cmd.AddCommand(newMessagesDeleteCmd(factory))
	cmd.AddCommand(newMessagesMarkReadCmd(factory))
	return cmd
}

// newMessagesConversationsCmd builds the "messages conversations" command.
func newMessagesConversationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conversations",
		Short: "List messaging conversations",
		RunE:  makeRunMessagesConversations(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of conversations to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunMessagesConversations(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// Step 1: resolve the current user's fsd_profile URN from /voyager/api/me.
		profileURN, err := resolveCurrentProfileURN(ctx, client)
		if err != nil {
			return fmt.Errorf("resolving current user profile: %w", err)
		}

		// Step 2: call the messaging GraphQL endpoint with the profile URN.
		variables := fmt.Sprintf("(mailboxUrn:%s)", profileURN)
		messagingPath := "/voyager/api/voyagerMessagingGraphQL/graphql"
		params := url.Values{}
		params.Set("queryId", messengerConversationsQueryID)
		params.Set("variables", variables)

		resp, err := client.Get(ctx, messagingPath, params)
		if err != nil {
			return fmt.Errorf("listing conversations: %w", err)
		}

		var normalized NormalizedResponse
		if err := client.DecodeJSON(resp, &normalized); err != nil {
			return fmt.Errorf("decoding conversations: %w", err)
		}

		summaries := extractConversationsFromGraphQL(normalized.Included)
		return printConversationSummaries(cmd, summaries)
	}
}

// resolveCurrentProfileURN calls /voyager/api/me and extracts the fsd_profile URN
// from the miniProfile entity in the included array.
func resolveCurrentProfileURN(ctx context.Context, client *Client) (string, error) {
	resp, err := client.Get(ctx, "/voyager/api/me", url.Values{})
	if err != nil {
		return "", fmt.Errorf("GET /voyager/api/me: %w", err)
	}

	var normalized NormalizedResponse
	if err := client.DecodeJSON(resp, &normalized); err != nil {
		return "", fmt.Errorf("decoding /voyager/api/me: %w", err)
	}

	// Look for a MiniProfile entity in included to get the entityUrn.
	raw := FindIncluded(normalized.Included, "MiniProfile")
	if raw != nil {
		var miniProfile struct {
			EntityURN string `json:"entityUrn"`
		}
		if err := json.Unmarshal(raw, &miniProfile); err == nil && miniProfile.EntityURN != "" {
			// Convert fs_miniProfile URN to fsd_profile URN for the messaging API.
			urn := strings.Replace(miniProfile.EntityURN, "urn:li:fs_miniProfile:", "urn:li:fsd_profile:", 1)
			return urn, nil
		}
	}

	// Fall back to the data field's *miniProfile reference if included lookup failed.
	var data struct {
		MiniProfile string `json:"*miniProfile"`
	}
	if err := json.Unmarshal(normalized.Data, &data); err == nil && data.MiniProfile != "" {
		urn := strings.Replace(data.MiniProfile, "urn:li:fs_miniProfile:", "urn:li:fsd_profile:", 1)
		return urn, nil
	}

	return "", fmt.Errorf("could not determine current user profile URN from /voyager/api/me")
}

// extractConversationsFromGraphQL finds all Conversation entities in the included array
// and maps them to ConversationSummary values.
func extractConversationsFromGraphQL(included []json.RawMessage) []ConversationSummary {
	rawEntities := FindAllIncluded(included, "com.linkedin.messenger.Conversation")
	summaries := make([]ConversationSummary, 0, len(rawEntities))
	for _, raw := range rawEntities {
		var entity messengerConversationEntity
		if err := json.Unmarshal(raw, &entity); err != nil {
			continue
		}
		summaries = append(summaries, ConversationSummary{
			ID:              entity.EntityURN,
			Title:           entity.Title,
			LastActivityAt:  entity.LastActivityAt,
			UnreadCount:     entity.UnreadCount,
			ParticipantURNs: []string{},
		})
	}
	return summaries
}

// newMessagesListCmd builds the "messages list" command.
func newMessagesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages in a conversation",
		RunE:  makeRunMessagesList(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	cmd.Flags().Int("limit", 20, "Maximum number of messages to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("conversation-id")
	return cmd
}

func makeRunMessagesList(factory ClientFactory) func(*cobra.Command, []string) error {
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
		params.Set("start", "0")
		if cursor != "" {
			params.Set("start", cursor)
		}
		params.Set("count", fmt.Sprintf("%d", limit))

		path := "/voyager/api/messaging/conversations/" + url.PathEscape(conversationID) + "/events"
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing messages in conversation %s: %w", conversationID, err)
		}

		var raw voyagerMessagesResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding messages: %w", err)
		}

		summaries := make([]MessageSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			summaries = append(summaries, toMessageSummary(el))
		}
		return printMessageSummaries(cmd, summaries)
	}
}

// newMessagesSendCmd builds the "messages send" command.
func newMessagesSendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message in an existing conversation",
		RunE:  makeRunMessagesSend(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	cmd.Flags().String("text", "", "Message text (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without sending")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("conversation-id")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunMessagesSend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		conversationID, _ := cmd.Flags().GetString("conversation-id")
		text, _ := cmd.Flags().GetString("text")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would send message to conversation %s", conversationID), map[string]any{
				"action":          "send",
				"conversation_id": conversationID,
				"text":            text,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"eventCreate": map[string]any{
				"value": map[string]any{
					"com.linkedin.voyager.messaging.create.MessageCreate": map[string]any{
						"body":        text,
						"attachments": []any{},
					},
				},
			},
		}

		path := "/voyager/api/messaging/conversations/" + url.PathEscape(conversationID) + "/events"
		resp, err := client.PostJSON(ctx, path, body)
		if err != nil {
			return fmt.Errorf("sending message to conversation %s: %w", conversationID, err)
		}

		var result map[string]any
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding send response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Message sent to conversation %s\n", conversationID)
		return nil
	}
}

// newMessagesNewCmd builds the "messages new" command.
func newMessagesNewCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new conversation and send a message",
		RunE:  makeRunMessagesNew(factory),
	}
	cmd.Flags().String("recipients", "", "Comma-separated list of recipient URNs (required)")
	cmd.Flags().String("text", "", "Message text (required)")
	cmd.Flags().String("subject", "", "Conversation subject (optional)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without sending")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("recipients")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunMessagesNew(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		recipientsStr, _ := cmd.Flags().GetString("recipients")
		text, _ := cmd.Flags().GetString("text")
		subject, _ := cmd.Flags().GetString("subject")

		recipients := strings.Split(recipientsStr, ",")
		for i, r := range recipients {
			recipients[i] = strings.TrimSpace(r)
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create new conversation with %s", recipientsStr), map[string]any{
				"action":     "new_conversation",
				"recipients": recipients,
				"text":       text,
				"subject":    subject,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		recipientList := make([]map[string]any, 0, len(recipients))
		for _, r := range recipients {
			recipientList = append(recipientList, map[string]any{
				"com.linkedin.voyager.messaging.MessagingMember": map[string]any{
					"miniProfile": map[string]any{
						"entityUrn": r,
					},
				},
			})
		}

		body := map[string]any{
			"keyVersion": "LEGACY_INBOX",
			"conversationCreate": map[string]any{
				"eventCreate": map[string]any{
					"value": map[string]any{
						"com.linkedin.voyager.messaging.create.MessageCreate": map[string]any{
							"body":        text,
							"attachments": []any{},
						},
					},
				},
				"recipients": recipientList,
				"subtype":    "MEMBER_TO_MEMBER",
			},
		}
		if subject != "" {
			convCreate := body["conversationCreate"].(map[string]any)
			convCreate["subject"] = subject
		}

		resp, err := client.PostJSON(ctx, "/voyager/api/messaging/conversations", body)
		if err != nil {
			return fmt.Errorf("creating new conversation: %w", err)
		}

		var result map[string]any
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding new conversation response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Println("New conversation created.")
		return nil
	}
}

// newMessagesDeleteCmd builds the "messages delete" command.
func newMessagesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a conversation",
		RunE:  makeRunMessagesDelete(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview the action without deleting")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("conversation-id")
	return cmd
}

func makeRunMessagesDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		conversationID, _ := cmd.Flags().GetString("conversation-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete conversation %s", conversationID), map[string]any{
				"action":          "delete",
				"conversation_id": conversationID,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/messaging/conversations/" + url.PathEscape(conversationID)
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("deleting conversation %s: %w", conversationID, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"deleted": true, "conversation_id": conversationID})
		}
		fmt.Printf("Deleted conversation %s\n", conversationID)
		return nil
	}
}

// newMessagesMarkReadCmd builds the "messages mark-read" command.
func newMessagesMarkReadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-read",
		Short: "Mark a conversation as read",
		RunE:  makeRunMessagesMarkRead(factory),
	}
	cmd.Flags().String("conversation-id", "", "Conversation ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without marking")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("conversation-id")
	return cmd
}

func makeRunMessagesMarkRead(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		conversationID, _ := cmd.Flags().GetString("conversation-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would mark conversation %s as read", conversationID), map[string]any{
				"action":          "mark_read",
				"conversation_id": conversationID,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/messaging/conversations/" + url.PathEscape(conversationID)
		resp, err := client.Patch(ctx, path, map[string]any{"patch": map[string]any{"read": true}})
		if err != nil {
			return fmt.Errorf("marking conversation %s as read: %w", conversationID, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"read": true, "conversation_id": conversationID})
		}
		fmt.Printf("Marked conversation %s as read\n", conversationID)
		return nil
	}
}

// printConversationSummaries outputs conversation summaries as JSON or text.
func printConversationSummaries(cmd *cobra.Command, convs []ConversationSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(convs)
	}
	if len(convs) == 0 {
		fmt.Println("No conversations found.")
		return nil
	}
	lines := make([]string, 0, len(convs)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-16s  %-7s", "ID", "PARTICIPANTS", "LAST ACTIVITY", "UNREAD"))
	for _, c := range convs {
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-16s  %-7d",
			truncate(c.ID, 20),
			truncate(c.Title, 40),
			formatTimestamp(c.LastActivityAt),
			c.UnreadCount,
		))
	}
	cli.PrintText(lines)
	return nil
}

// printMessageSummaries outputs message summaries as JSON or text.
func printMessageSummaries(cmd *cobra.Command, msgs []MessageSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(msgs)
	}
	if len(msgs) == 0 {
		fmt.Println("No messages found.")
		return nil
	}
	lines := make([]string, 0, len(msgs)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-16s  %-50s", "SENDER", "DATE", "TEXT"))
	for _, m := range msgs {
		name := m.SenderName
		if name == "" {
			name = truncate(m.SenderURN, 20)
		}
		lines = append(lines, fmt.Sprintf("%-20s  %-16s  %-50s",
			truncate(name, 20),
			formatTimestamp(m.Timestamp),
			truncate(m.Text, 50),
		))
	}
	cli.PrintText(lines)
	return nil
}
