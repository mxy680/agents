package imessage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newScheduledCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scheduled",
		Short:   "Manage scheduled messages",
		Aliases: []string{"sched"},
	}

	cmd.AddCommand(newScheduledListCmd(factory))
	cmd.AddCommand(newScheduledGetCmd(factory))
	cmd.AddCommand(newScheduledCreateCmd(factory))
	cmd.AddCommand(newScheduledUpdateCmd(factory))
	cmd.AddCommand(newScheduledDeleteCmd(factory))

	return cmd
}

// newScheduledListCmd lists all scheduled messages.
func newScheduledListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all scheduled messages",
		RunE:  makeRunScheduledList(factory),
	}
	return cmd
}

func makeRunScheduledList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Get(ctx, "message/schedule", nil)
		if err != nil {
			return fmt.Errorf("listing scheduled messages: %w", err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		var items []json.RawMessage
		if err := json.Unmarshal(data, &items); err != nil {
			// Tolerate empty or unexpected payload.
			return printResult(cmd, []ScheduledMessageSummary{}, []string{"No scheduled messages found"})
		}

		summaries := make([]ScheduledMessageSummary, 0, len(items))
		for _, item := range items {
			summaries = append(summaries, toScheduledSummary(item))
		}

		lines := make([]string, 0, len(summaries))
		for _, s := range summaries {
			lines = append(lines, formatScheduledLine(s))
		}
		if len(lines) == 0 {
			lines = []string{"No scheduled messages found"}
		}
		return printResult(cmd, summaries, lines)
	}
}

// newScheduledGetCmd retrieves a single scheduled message by ID.
func newScheduledGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a scheduled message by ID",
		RunE:  makeRunScheduledGet(factory),
	}
	cmd.Flags().Int("id", 0, "Scheduled message ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunScheduledGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetInt("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Get(ctx, fmt.Sprintf("message/schedule/%d", id), nil)
		if err != nil {
			return fmt.Errorf("getting scheduled message %d: %w", id, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		s := toScheduledSummary(data)
		return printResult(cmd, s, []string{
			fmt.Sprintf("ID:        %d", s.ID),
			fmt.Sprintf("Chat:      %s", s.ChatGUID),
			fmt.Sprintf("Message:   %s", truncate(s.Message, 60)),
			fmt.Sprintf("Send At:   %s", formatTimestamp(s.SendDate)),
			fmt.Sprintf("Status:    %s", s.Status),
		})
	}
}

// newScheduledCreateCmd creates a new scheduled message.
func newScheduledCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Schedule a message to be sent at a specific time",
		RunE:  makeRunScheduledCreate(factory),
	}
	cmd.Flags().String("chat-guid", "", "Chat GUID to send the message to (required)")
	cmd.Flags().String("text", "", "Message text (required)")
	cmd.Flags().String("send-at", "", "Send time in RFC3339 format, e.g. 2026-01-15T14:30:00Z (required)")
	_ = cmd.MarkFlagRequired("chat-guid")
	_ = cmd.MarkFlagRequired("text")
	_ = cmd.MarkFlagRequired("send-at")
	return cmd
}

func makeRunScheduledCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		chatGUID, _ := cmd.Flags().GetString("chat-guid")
		text, _ := cmd.Flags().GetString("text")
		sendAt, _ := cmd.Flags().GetString("send-at")

		t, err := time.Parse(time.RFC3339, sendAt)
		if err != nil {
			return fmt.Errorf("invalid --send-at value %q; expected RFC3339, e.g. 2026-01-15T14:30:00Z: %w", sendAt, err)
		}
		scheduledFor := t.UnixMilli()

		body := map[string]any{
			"chatGuid":     chatGUID,
			"message":      text,
			"scheduledFor": scheduledFor,
		}

		if isDryRun(cmd) {
			result := dryRunResult("create scheduled message", map[string]any{
				"chat_guid":     chatGUID,
				"message":       text,
				"scheduled_for": scheduledFor,
			})
			return printResult(cmd, result, []string{
				fmt.Sprintf("[dry-run] Would schedule message to %s at %s: %s", chatGUID, t.Format("2006-01-02 15:04:05"), truncate(text, 60)),
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Post(ctx, "message/schedule", body)
		if err != nil {
			return fmt.Errorf("creating scheduled message: %w", err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		s := toScheduledSummary(data)
		return printResult(cmd, s, []string{
			fmt.Sprintf("Scheduled message created"),
			fmt.Sprintf("ID:      %d", s.ID),
			fmt.Sprintf("Chat:    %s", s.ChatGUID),
			fmt.Sprintf("Send At: %s", formatTimestamp(s.SendDate)),
			fmt.Sprintf("Message: %s", truncate(s.Message, 60)),
		})
	}
}

// newScheduledUpdateCmd updates an existing scheduled message.
func newScheduledUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a scheduled message",
		RunE:  makeRunScheduledUpdate(factory),
	}
	cmd.Flags().Int("id", 0, "Scheduled message ID (required)")
	cmd.Flags().String("text", "", "New message text")
	cmd.Flags().String("send-at", "", "New send time in RFC3339 format")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunScheduledUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetInt("id")
		text, _ := cmd.Flags().GetString("text")
		sendAt, _ := cmd.Flags().GetString("send-at")

		if text == "" && sendAt == "" {
			return fmt.Errorf("at least one of --text or --send-at must be provided")
		}

		body := map[string]any{}
		if text != "" {
			body["message"] = text
		}
		if sendAt != "" {
			t, err := time.Parse(time.RFC3339, sendAt)
			if err != nil {
				return fmt.Errorf("invalid --send-at value %q; expected RFC3339, e.g. 2026-01-15T14:30:00Z: %w", sendAt, err)
			}
			body["scheduledFor"] = t.UnixMilli()
		}

		if isDryRun(cmd) {
			result := dryRunResult("update scheduled message", map[string]any{
				"id":   id,
				"body": body,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would update scheduled message %d", id)})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Put(ctx, fmt.Sprintf("message/schedule/%d", id), body)
		if err != nil {
			return fmt.Errorf("updating scheduled message %d: %w", id, err)
		}
		data, err := ParseResponse(raw)
		if err != nil {
			return err
		}

		s := toScheduledSummary(data)
		return printResult(cmd, s, []string{
			fmt.Sprintf("Updated scheduled message %d", id),
			fmt.Sprintf("Send At: %s", formatTimestamp(s.SendDate)),
			fmt.Sprintf("Message: %s", truncate(s.Message, 60)),
		})
	}
}

// newScheduledDeleteCmd deletes a scheduled message.
func newScheduledDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a scheduled message",
		RunE:  makeRunScheduledDelete(factory),
	}
	cmd.Flags().Int("id", 0, "Scheduled message ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive delete action")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunScheduledDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetInt("id")

		if isDryRun(cmd) {
			result := dryRunResult("delete scheduled message", map[string]any{"id": id})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would delete scheduled message %d", id)})
		}

		if err := confirmDestructive(cmd, "delete scheduled message"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := client.Delete(ctx, fmt.Sprintf("message/schedule/%d", id))
		if err != nil {
			return fmt.Errorf("deleting scheduled message %d: %w", id, err)
		}
		if _, err := ParseResponse(raw); err != nil {
			return err
		}
		return printResult(cmd, map[string]any{"id": id, "deleted": true}, []string{
			fmt.Sprintf("Deleted scheduled message %d", id),
		})
	}
}

// toScheduledSummary converts a raw JSON scheduled message to a ScheduledMessageSummary.
func toScheduledSummary(raw json.RawMessage) ScheduledMessageSummary {
	var m map[string]any
	json.Unmarshal(raw, &m)
	return ScheduledMessageSummary{
		ID:       int(getInt64(m, "id")),
		ChatGUID: getString(m, "chatGuid"),
		Message:  getString(m, "message"),
		SendDate: getInt64(m, "scheduledFor"),
		Status:   getString(m, "status"),
	}
}

// formatScheduledLine formats a scheduled message for text output.
func formatScheduledLine(s ScheduledMessageSummary) string {
	status := s.Status
	if status == "" {
		status = "pending"
	}
	return fmt.Sprintf("[%d] %-20s  %s  %s",
		s.ID,
		truncate(s.ChatGUID, 18),
		formatTimestamp(s.SendDate),
		truncate(s.Message, 40),
	)
}
