package imessage

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newMessagesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messages",
		Short:   "Send and manage messages",
		Aliases: []string{"msg"},
	}

	cmd.AddCommand(newMessagesSendCmd(factory))
	cmd.AddCommand(newMessagesSendGroupCmd(factory))
	cmd.AddCommand(newMessagesSendAttachmentCmd(factory))
	cmd.AddCommand(newMessagesSendMultipartCmd(factory))
	cmd.AddCommand(newMessagesGetCmd(factory))
	cmd.AddCommand(newMessagesQueryCmd(factory))
	cmd.AddCommand(newMessagesEditCmd(factory))
	cmd.AddCommand(newMessagesUnsendCmd(factory))
	cmd.AddCommand(newMessagesReactCmd(factory))
	cmd.AddCommand(newMessagesDeleteCmd(factory))
	cmd.AddCommand(newMessagesCountCmd(factory))
	cmd.AddCommand(newMessagesCountUpdatedCmd(factory))
	cmd.AddCommand(newMessagesCountSentCmd(factory))
	cmd.AddCommand(newMessagesEmbeddedMediaCmd(factory))
	cmd.AddCommand(newMessagesNotifyCmd(factory))

	return cmd
}

// newMessagesSendCmd sends a message to an individual contact via their phone/email.
func newMessagesSendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message to a phone number or email address",
		RunE:  makeRunMessagesSend(factory),
	}
	cmd.Flags().String("to", "", "Recipient phone number or email address (required)")
	cmd.Flags().String("text", "", "Message text (required)")
	cmd.Flags().String("subject", "", "Message subject")
	cmd.Flags().String("effect", "", "iMessage effect identifier")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunMessagesSend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		to, _ := cmd.Flags().GetString("to")
		text, _ := cmd.Flags().GetString("text")
		subject, _ := cmd.Flags().GetString("subject")
		effect, _ := cmd.Flags().GetString("effect")

		chatGUID := "iMessage;-;" + to
		body := map[string]any{
			"chatGuid": chatGUID,
			"message":  text,
			"method":   "apple-script",
		}
		if subject != "" {
			body["subject"] = subject
		}
		if effect != "" {
			body["effectId"] = effect
		}

		if isDryRun(cmd) {
			result := dryRunResult("send message", map[string]any{
				"chat_guid": chatGUID,
				"message":   text,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would send message to %s: %s", to, truncate(text, 60))})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, "message/text", body)
		if err != nil {
			return fmt.Errorf("sending message to %s: %w", to, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}
		msg := toMessageSummary(data)
		return printResult(cmd, msg, []string{
			fmt.Sprintf("Sent:  %s", truncate(msg.Text, 60)),
			fmt.Sprintf("GUID:  %s", msg.GUID),
		})
	}
}

// newMessagesSendGroupCmd sends a message to an existing group chat by GUID.
func newMessagesSendGroupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-group",
		Short: "Send a message to a group chat by GUID",
		RunE:  makeRunMessagesSendGroup(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	cmd.Flags().String("text", "", "Message text (required)")
	cmd.Flags().String("subject", "", "Message subject")
	cmd.Flags().String("effect", "", "iMessage effect identifier")
	_ = cmd.MarkFlagRequired("guid")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunMessagesSendGroup(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		text, _ := cmd.Flags().GetString("text")
		subject, _ := cmd.Flags().GetString("subject")
		effect, _ := cmd.Flags().GetString("effect")

		body := map[string]any{
			"chatGuid": guid,
			"message":  text,
		}
		if subject != "" {
			body["subject"] = subject
		}
		if effect != "" {
			body["effectId"] = effect
		}

		if isDryRun(cmd) {
			result := dryRunResult("send group message", map[string]any{
				"chat_guid": guid,
				"message":   text,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would send message to group %s: %s", guid, truncate(text, 60))})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, "message/text", body)
		if err != nil {
			return fmt.Errorf("sending message to group %s: %w", guid, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}
		msg := toMessageSummary(data)
		return printResult(cmd, msg, []string{
			fmt.Sprintf("Sent:  %s", truncate(msg.Text, 60)),
			fmt.Sprintf("GUID:  %s", msg.GUID),
		})
	}
}

// newMessagesSendAttachmentCmd sends a file attachment to a contact.
// TODO: Full multipart/form-data upload is not yet implemented; only text fallback is sent.
func newMessagesSendAttachmentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-attachment",
		Short: "Send an attachment to a phone number or email address",
		RunE:  makeRunMessagesSendAttachment(factory),
	}
	cmd.Flags().String("to", "", "Recipient phone number or email address (required)")
	cmd.Flags().String("path", "", "Local file path to attach (required)")
	cmd.Flags().String("text", "", "Optional message text to accompany the attachment")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func makeRunMessagesSendAttachment(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		to, _ := cmd.Flags().GetString("to")
		path, _ := cmd.Flags().GetString("path")
		text, _ := cmd.Flags().GetString("text")

		chatGUID := "iMessage;-;" + to

		if isDryRun(cmd) {
			result := dryRunResult("send attachment", map[string]any{
				"chat_guid":  chatGUID,
				"attachment": path,
				"text":       text,
			})
			return printResult(cmd, result, []string{
				fmt.Sprintf("[dry-run] Would send attachment %s to %s", path, to),
			})
		}

		// TODO: Implement full multipart/form-data upload to POST /api/v1/message/attachment.
		// For now, verify the file exists and send any accompanying text.
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("attachment file not found: %w", err)
		}

		if text == "" {
			return fmt.Errorf("full attachment upload not yet implemented; provide --text to send a text message instead")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"chatGuid": chatGUID,
			"message":  text,
			"method":   "apple-script",
		}
		raw, err := client.Post(ctx, "message/text", body)
		if err != nil {
			return fmt.Errorf("sending message to %s: %w", to, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}
		msg := toMessageSummary(data)
		return printResult(cmd, msg, []string{
			fmt.Sprintf("Sent:  %s", truncate(msg.Text, 60)),
			fmt.Sprintf("GUID:  %s", msg.GUID),
			fmt.Sprintf("Note:  attachment upload not yet implemented; text sent only"),
		})
	}
}

// newMessagesSendMultipartCmd sends a multipart message using the BlueBubbles multipart API.
func newMessagesSendMultipartCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-multipart",
		Short: "Send a multipart message (text + attachments) to a recipient",
		RunE:  makeRunMessagesSendMultipart(factory),
	}
	cmd.Flags().String("to", "", "Recipient phone number or email address (required)")
	cmd.Flags().String("parts", "", "JSON array of message parts")
	cmd.Flags().String("parts-file", "", "Path to JSON file containing message parts")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func makeRunMessagesSendMultipart(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		to, _ := cmd.Flags().GetString("to")
		partsStr, _ := cmd.Flags().GetString("parts")
		partsFile, _ := cmd.Flags().GetString("parts-file")

		if partsStr == "" && partsFile == "" {
			return fmt.Errorf("one of --parts or --parts-file is required")
		}

		var partsJSON json.RawMessage
		if partsFile != "" {
			data, err := os.ReadFile(partsFile)
			if err != nil {
				return fmt.Errorf("reading parts file: %w", err)
			}
			partsJSON = json.RawMessage(data)
		} else {
			partsJSON = json.RawMessage(partsStr)
		}
		if !json.Valid(partsJSON) {
			return fmt.Errorf("parts is not valid JSON")
		}

		chatGUID := "iMessage;-;" + to

		if isDryRun(cmd) {
			result := dryRunResult("send multipart message", map[string]any{
				"chat_guid": chatGUID,
				"parts":     partsJSON,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would send multipart message to %s", to)})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"chatGuid": chatGUID,
			"parts":    partsJSON,
		}
		raw, err := client.Post(ctx, "message/multipart", body)
		if err != nil {
			return fmt.Errorf("sending multipart message to %s: %w", to, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}
		msg := toMessageSummary(data)
		return printResult(cmd, msg, []string{
			fmt.Sprintf("Sent multipart message"),
			fmt.Sprintf("GUID:  %s", msg.GUID),
		})
	}
}

// newMessagesGetCmd retrieves a single message by GUID.
func newMessagesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a message by GUID",
		RunE:  makeRunMessagesGet(factory),
	}
	cmd.Flags().String("guid", "", "Message GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunMessagesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Get(ctx, fmt.Sprintf("message/%s", guid), nil)
		if err != nil {
			return fmt.Errorf("getting message %s: %w", guid, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}
		msg := toMessageSummary(data)
		return printResult(cmd, msg, []string{
			fmt.Sprintf("GUID:    %s", msg.GUID),
			fmt.Sprintf("From me: %v", msg.IsFromMe),
			fmt.Sprintf("Handle:  %s", msg.Handle),
			fmt.Sprintf("Date:    %s", formatTimestamp(msg.DateCreated)),
			fmt.Sprintf("Text:    %s", msg.Text),
		})
	}
}

// newMessagesQueryCmd queries messages with optional filters.
func newMessagesQueryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query messages with optional filters",
		RunE:  makeRunMessagesQuery(factory),
	}
	cmd.Flags().String("chat-guid", "", "Filter by chat GUID")
	cmd.Flags().Int("limit", 25, "Maximum number of messages to return")
	cmd.Flags().Int("offset", 0, "Pagination offset")
	cmd.Flags().String("after", "", "Return messages after this timestamp (ms since epoch)")
	cmd.Flags().String("before", "", "Return messages before this timestamp (ms since epoch)")
	cmd.Flags().String("sort", "DESC", "Sort order: ASC or DESC")
	cmd.Flags().String("with", "", "Comma-separated list of relations to include (e.g. chat,attachment,handle)")
	return cmd
}

func makeRunMessagesQuery(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		chatGUID, _ := cmd.Flags().GetString("chat-guid")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		after, _ := cmd.Flags().GetString("after")
		before, _ := cmd.Flags().GetString("before")
		sort, _ := cmd.Flags().GetString("sort")
		with, _ := cmd.Flags().GetString("with")

		body := map[string]any{
			"limit":  limit,
			"offset": offset,
			"sort":   sort,
		}
		if chatGUID != "" {
			body["chatGuid"] = chatGUID
		}
		if after != "" {
			body["after"] = after
		}
		if before != "" {
			body["before"] = before
		}
		if with != "" {
			body["with"] = strings.Split(with, ",")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, "message/query", body)
		if err != nil {
			return fmt.Errorf("querying messages: %w", err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		var msgs []json.RawMessage
		if err := json.Unmarshal(data, &msgs); err != nil {
			// data may be a single object or wrapped; try extracting messages key
			var wrapper map[string]json.RawMessage
			if err2 := json.Unmarshal(data, &wrapper); err2 == nil {
				if m, ok := wrapper["messages"]; ok {
					json.Unmarshal(m, &msgs)
				}
			}
		}

		summaries := make([]MessageSummary, 0, len(msgs))
		for _, m := range msgs {
			summaries = append(summaries, toMessageSummary(m))
		}

		lines := make([]string, 0, len(summaries))
		for _, m := range summaries {
			lines = append(lines, formatMessageLine(m))
		}
		return printResult(cmd, summaries, lines)
	}
}

// newMessagesEditCmd edits a previously sent message.
func newMessagesEditCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a previously sent message",
		RunE:  makeRunMessagesEdit(factory),
	}
	cmd.Flags().String("guid", "", "Message GUID (required)")
	cmd.Flags().String("text", "", "New message text (required)")
	cmd.Flags().Int("part", 0, "Part index to edit (default 0)")
	_ = cmd.MarkFlagRequired("guid")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}

func makeRunMessagesEdit(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		text, _ := cmd.Flags().GetString("text")
		part, _ := cmd.Flags().GetInt("part")

		body := map[string]any{
			"editedMessage": text,
			"partIndex":     part,
		}

		if isDryRun(cmd) {
			result := dryRunResult("edit message", map[string]any{
				"guid":   guid,
				"text":   text,
				"part":   part,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would edit message %s (part %d): %s", guid, part, truncate(text, 60))})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, fmt.Sprintf("message/%s/edit", guid), body)
		if err != nil {
			return fmt.Errorf("editing message %s: %w", guid, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}
		msg := toMessageSummary(data)
		return printResult(cmd, msg, []string{
			fmt.Sprintf("Edited message %s", guid),
			fmt.Sprintf("Text:  %s", truncate(msg.Text, 60)),
		})
	}
}

// newMessagesUnsendCmd unsends (deletes) a message part.
func newMessagesUnsendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsend",
		Short: "Unsend a message or message part",
		RunE:  makeRunMessagesUnsend(factory),
	}
	cmd.Flags().String("guid", "", "Message GUID (required)")
	cmd.Flags().Int("part", 0, "Part index to unsend (default 0)")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunMessagesUnsend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		part, _ := cmd.Flags().GetInt("part")

		body := map[string]any{
			"partIndex": part,
		}

		if isDryRun(cmd) {
			result := dryRunResult("unsend message", map[string]any{
				"guid": guid,
				"part": part,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would unsend message %s (part %d)", guid, part)})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, fmt.Sprintf("message/%s/unsend", guid), body)
		if err != nil {
			return fmt.Errorf("unsending message %s: %w", guid, err)
		}
		if _, err := ParseResponse(raw); err != nil {
			return err
		}
		return printResult(cmd, map[string]any{"guid": guid, "unsent": true}, []string{
			fmt.Sprintf("Unsent message %s (part %d)", guid, part),
		})
	}
}

// newMessagesReactCmd adds a tapback reaction to a message.
func newMessagesReactCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "react",
		Short: "Add a tapback reaction to a message",
		RunE:  makeRunMessagesReact(factory),
	}
	cmd.Flags().String("chat-guid", "", "Chat GUID (required)")
	cmd.Flags().String("message-guid", "", "Message GUID to react to (required)")
	cmd.Flags().String("type", "", "Reaction type: love, like, dislike, laugh, emphasis, question (required)")
	_ = cmd.MarkFlagRequired("chat-guid")
	_ = cmd.MarkFlagRequired("message-guid")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func makeRunMessagesReact(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		chatGUID, _ := cmd.Flags().GetString("chat-guid")
		messageGUID, _ := cmd.Flags().GetString("message-guid")
		reactionType, _ := cmd.Flags().GetString("type")

		validReactions := map[string]bool{
			"love": true, "like": true, "dislike": true,
			"laugh": true, "emphasis": true, "question": true,
		}
		if !validReactions[reactionType] {
			return fmt.Errorf("invalid reaction type %q; must be one of: love, like, dislike, laugh, emphasis, question", reactionType)
		}

		body := map[string]any{
			"chatGuid":            chatGUID,
			"selectedMessageGuid": messageGUID,
			"reaction":            reactionType,
		}

		if isDryRun(cmd) {
			result := dryRunResult("react to message", map[string]any{
				"chat_guid":    chatGUID,
				"message_guid": messageGUID,
				"reaction":     reactionType,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would react with %q to message %s", reactionType, messageGUID)})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, "message/react", body)
		if err != nil {
			return fmt.Errorf("reacting to message %s: %w", messageGUID, err)
		}
		if _, err := ParseResponse(raw); err != nil {
			return err
		}
		return printResult(cmd, map[string]any{
			"chat_guid":    chatGUID,
			"message_guid": messageGUID,
			"reaction":     reactionType,
		}, []string{
			fmt.Sprintf("Reacted with %q to message %s", reactionType, messageGUID),
		})
	}
}

// newMessagesDeleteCmd deletes a message from a chat.
func newMessagesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a message from a chat",
		RunE:  makeRunMessagesDelete(factory),
	}
	cmd.Flags().String("chat-guid", "", "Chat GUID (required)")
	cmd.Flags().String("message-guid", "", "Message GUID to delete (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive delete action")
	_ = cmd.MarkFlagRequired("chat-guid")
	_ = cmd.MarkFlagRequired("message-guid")
	return cmd
}

func makeRunMessagesDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		chatGUID, _ := cmd.Flags().GetString("chat-guid")
		messageGUID, _ := cmd.Flags().GetString("message-guid")

		if isDryRun(cmd) {
			result := dryRunResult("delete message", map[string]any{
				"chat_guid":    chatGUID,
				"message_guid": messageGUID,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would delete message %s from chat %s", messageGUID, chatGUID)})
		}

		if err := confirmDestructive(cmd, "delete message"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Delete(ctx, fmt.Sprintf("chat/%s/%s", chatGUID, messageGUID))
		if err != nil {
			return fmt.Errorf("deleting message %s from chat %s: %w", messageGUID, chatGUID, err)
		}
		if _, err := ParseResponse(raw); err != nil {
			return err
		}
		return printResult(cmd, map[string]any{
			"chat_guid":    chatGUID,
			"message_guid": messageGUID,
			"deleted":      true,
		}, []string{
			fmt.Sprintf("Deleted message %s from chat %s", messageGUID, chatGUID),
		})
	}
}

// newMessagesCountCmd returns total message count with optional filters.
func newMessagesCountCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get total message count",
		RunE:  makeRunMessagesCount(factory),
	}
	cmd.Flags().String("after", "", "Count messages after this timestamp (ms since epoch)")
	cmd.Flags().String("before", "", "Count messages before this timestamp (ms since epoch)")
	cmd.Flags().String("chat-guid", "", "Filter by chat GUID")
	return cmd
}

func makeRunMessagesCount(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		after, _ := cmd.Flags().GetString("after")
		before, _ := cmd.Flags().GetString("before")
		chatGUID, _ := cmd.Flags().GetString("chat-guid")

		params := url.Values{}
		if after != "" {
			params.Set("after", after)
		}
		if before != "" {
			params.Set("before", before)
		}
		if chatGUID != "" {
			params.Set("chatGuid", chatGUID)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Get(ctx, "message/count", params)
		if err != nil {
			return fmt.Errorf("getting message count: %w", err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		var result map[string]any
		json.Unmarshal(data, &result)
		return printResult(cmd, result, []string{fmt.Sprintf("Message count: %v", result["total"])})
	}
}

// newMessagesCountUpdatedCmd returns count of updated messages.
func newMessagesCountUpdatedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count-updated",
		Short: "Get count of updated messages",
		RunE:  makeRunMessagesCountUpdated(factory),
	}
	cmd.Flags().String("after", "", "Count messages updated after this timestamp (ms since epoch)")
	cmd.Flags().String("before", "", "Count messages updated before this timestamp (ms since epoch)")
	cmd.Flags().String("chat-guid", "", "Filter by chat GUID")
	return cmd
}

func makeRunMessagesCountUpdated(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		after, _ := cmd.Flags().GetString("after")
		before, _ := cmd.Flags().GetString("before")
		chatGUID, _ := cmd.Flags().GetString("chat-guid")

		params := url.Values{}
		if after != "" {
			params.Set("after", after)
		}
		if before != "" {
			params.Set("before", before)
		}
		if chatGUID != "" {
			params.Set("chatGuid", chatGUID)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Get(ctx, "message/count/updated", params)
		if err != nil {
			return fmt.Errorf("getting updated message count: %w", err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		var result map[string]any
		json.Unmarshal(data, &result)
		return printResult(cmd, result, []string{fmt.Sprintf("Updated message count: %v", result["total"])})
	}
}

// newMessagesCountSentCmd returns count of messages sent by the current user.
func newMessagesCountSentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count-sent",
		Short: "Get count of messages sent by me",
		RunE:  makeRunMessagesCountSent(factory),
	}
	return cmd
}

func makeRunMessagesCountSent(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Get(ctx, "message/count/me", nil)
		if err != nil {
			return fmt.Errorf("getting sent message count: %w", err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		var result map[string]any
		json.Unmarshal(data, &result)
		return printResult(cmd, result, []string{fmt.Sprintf("Sent message count: %v", result["total"])})
	}
}

// newMessagesEmbeddedMediaCmd retrieves embedded media for a message.
func newMessagesEmbeddedMediaCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "embedded-media",
		Short: "Get embedded media for a message",
		RunE:  makeRunMessagesEmbeddedMedia(factory),
	}
	cmd.Flags().String("guid", "", "Message GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunMessagesEmbeddedMedia(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Get(ctx, fmt.Sprintf("message/%s/embedded-media", guid), nil)
		if err != nil {
			return fmt.Errorf("getting embedded media for message %s: %w", guid, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		var media []json.RawMessage
		json.Unmarshal(data, &media)

		summaries := make([]AttachmentSummary, 0, len(media))
		for _, m := range media {
			summaries = append(summaries, toAttachmentSummary(m))
		}

		lines := make([]string, 0, len(summaries))
		for _, a := range summaries {
			lines = append(lines, fmt.Sprintf("%-40s  %s  %d bytes", truncate(a.FileName, 38), a.MIMEType, a.TotalBytes))
		}
		if len(lines) == 0 {
			lines = []string{"No embedded media found"}
		}
		return printResult(cmd, summaries, lines)
	}
}

// newMessagesNotifyCmd triggers a notification for a message.
func newMessagesNotifyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Trigger a notification for a message",
		RunE:  makeRunMessagesNotify(factory),
	}
	cmd.Flags().String("guid", "", "Message GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunMessagesNotify(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		if isDryRun(cmd) {
			result := dryRunResult("notify message", map[string]any{"guid": guid})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would trigger notification for message %s", guid)})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, fmt.Sprintf("message/%s/notify", guid), nil)
		if err != nil {
			return fmt.Errorf("notifying message %s: %w", guid, err)
		}
		if _, err := ParseResponse(raw); err != nil {
			return err
		}
		return printResult(cmd, map[string]any{"guid": guid, "notified": true}, []string{
			fmt.Sprintf("Triggered notification for message %s", guid),
		})
	}
}

// isDryRun is a local helper that delegates to the root persistent flag.
func isDryRun(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("dry-run")
	return v
}
