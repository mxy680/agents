package imessage

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newChatsCmd returns the parent "chats" command with all subcommands attached.
func newChatsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chats",
		Short:   "Manage iMessage chats",
		Aliases: []string{"chat"},
	}

	cmd.AddCommand(newChatsListCmd(factory))
	cmd.AddCommand(newChatsGetCmd(factory))
	cmd.AddCommand(newChatsCreateCmd(factory))
	cmd.AddCommand(newChatsUpdateCmd(factory))
	cmd.AddCommand(newChatsDeleteCmd(factory))
	cmd.AddCommand(newChatsMessagesCmd(factory))
	cmd.AddCommand(newChatsReadCmd(factory))
	cmd.AddCommand(newChatsUnreadCmd(factory))
	cmd.AddCommand(newChatsLeaveCmd(factory))
	cmd.AddCommand(newChatsTypingCmd(factory))
	cmd.AddCommand(newChatsCountCmd(factory))
	cmd.AddCommand(newChatsIconCmd(factory))

	return cmd
}

// --- chats list ---

func newChatsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List chats",
		RunE:  makeRunChatsList(factory),
	}
	cmd.Flags().Int("limit", 25, "Maximum number of chats to return")
	cmd.Flags().Int("offset", 0, "Offset for pagination")
	cmd.Flags().Bool("with-participants", true, "Include participants in response")
	cmd.Flags().Bool("with-last-message", true, "Include last message in response")
	cmd.Flags().String("sort", "lastmessage", "Sort order (lastmessage, etc.)")
	cmd.Flags().String("query", "", "Filter chats by display name or identifier")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		withParticipants, _ := cmd.Flags().GetBool("with-participants")
		withLastMessage, _ := cmd.Flags().GetBool("with-last-message")
		sort, _ := cmd.Flags().GetString("sort")
		query, _ := cmd.Flags().GetString("query")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		with := []string{}
		if withParticipants {
			with = append(with, "participants")
		}
		if withLastMessage {
			with = append(with, "lastMessage")
		}

		reqBody := map[string]any{
			"limit":  limit,
			"offset": offset,
			"with":   with,
			"sort":   sort,
		}
		if query != "" {
			reqBody["where"] = []map[string]any{
				{"statement": map[string]any{"name": "chatIdentifier", "operator": "LIKE", "value": "%" + query + "%"}},
			}
		}

		body, err := client.Post(ctx, "chat/query", reqBody)
		if err != nil {
			return fmt.Errorf("querying chats: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var rawChats []json.RawMessage
		if err := json.Unmarshal(data, &rawChats); err != nil {
			// Some versions wrap in {data: [...]}
			var wrapper struct {
				Data []json.RawMessage `json:"data"`
			}
			if err2 := json.Unmarshal(data, &wrapper); err2 != nil {
				return fmt.Errorf("parse chats: %w", err)
			}
			rawChats = wrapper.Data
		}

		summaries := make([]ChatSummary, 0, len(rawChats))
		for _, r := range rawChats {
			summaries = append(summaries, toChatSummary(r))
		}

		var lines []string
		for _, c := range summaries {
			lines = append(lines, formatChatLine(c))
		}

		return printResult(cmd, summaries, lines)
	}
}

// --- chats get ---

func newChatsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a chat by GUID",
		RunE:  makeRunChatsGet(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, "chat/"+guid, nil)
		if err != nil {
			return fmt.Errorf("getting chat %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toChatSummary(data)
		lines := []string{formatChatLine(summary)}
		if len(summary.Participants) > 0 {
			lines = append(lines, "  Participants: "+strings.Join(summary.Participants, ", "))
		}

		return printResult(cmd, summary, lines)
	}
}

// --- chats create ---

func newChatsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new chat",
		RunE:  makeRunChatsCreate(factory),
	}
	cmd.Flags().StringSlice("participants", nil, "Comma-separated participant addresses (required)")
	_ = cmd.MarkFlagRequired("participants")
	cmd.Flags().String("message", "", "Initial message to send")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		participants, _ := cmd.Flags().GetStringSlice("participants")
		message, _ := cmd.Flags().GetString("message")

		if cli.IsDryRun(cmd) {
			details := map[string]any{"participants": participants}
			if message != "" {
				details["message"] = message
			}
			return printResult(cmd, dryRunResult("create chat", details), []string{"[dry-run] would create chat with: " + strings.Join(participants, ", ")})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"participants": participants,
			"service":      "iMessage",
		}
		if message != "" {
			reqBody["message"] = message
		}

		body, err := client.Post(ctx, "chat/new", reqBody)
		if err != nil {
			return fmt.Errorf("creating chat: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toChatSummary(data)
		lines := []string{"Created chat: " + summary.GUID}
		if len(summary.Participants) > 0 {
			lines = append(lines, "  Participants: "+strings.Join(summary.Participants, ", "))
		}

		return printResult(cmd, summary, lines)
	}
}

// --- chats update ---

func newChatsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a chat's display name",
		RunE:  makeRunChatsUpdate(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().String("name", "", "New display name")
	cmd.Flags().Bool("dry-run", false, "Preview without updating")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("update chat", map[string]any{"guid": guid, "displayName": name}),
				[]string{fmt.Sprintf("[dry-run] would update chat %s: displayName=%q", guid, name)},
			)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"displayName": name,
		}

		body, err := client.Put(ctx, "chat/"+guid, reqBody)
		if err != nil {
			return fmt.Errorf("updating chat %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toChatSummary(data)
		return printResult(cmd, summary, []string{"Updated chat: " + summary.GUID})
	}
}

// --- chats delete ---

func newChatsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a chat",
		RunE:  makeRunChatsDelete(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	cmd.Flags().Bool("dry-run", false, "Preview without deleting")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("delete chat", map[string]any{"guid": guid}),
				[]string{fmt.Sprintf("[dry-run] would delete chat %s", guid)},
			)
		}

		if err := confirmDestructive(cmd, "delete chat"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Delete(ctx, "chat/"+guid)
		if err != nil {
			return fmt.Errorf("deleting chat %s: %w", guid, err)
		}

		_, err = ParseResponse(body)
		if err != nil {
			return err
		}

		return printResult(cmd, map[string]any{"guid": guid, "deleted": true}, []string{"Deleted chat: " + guid})
	}
}

// --- chats messages ---

func newChatsMessagesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "List messages in a chat",
		RunE:  makeRunChatsMessages(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Int("limit", 25, "Maximum number of messages to return")
	cmd.Flags().Int("offset", 0, "Offset for pagination")
	cmd.Flags().Int64("after", 0, "Return messages after this timestamp (ms)")
	cmd.Flags().Int64("before", 0, "Return messages before this timestamp (ms)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsMessages(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		after, _ := cmd.Flags().GetInt64("after")
		before, _ := cmd.Flags().GetInt64("before")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))
		params.Set("offset", strconv.Itoa(offset))
		if after > 0 {
			params.Set("after", strconv.FormatInt(after, 10))
		}
		if before > 0 {
			params.Set("before", strconv.FormatInt(before, 10))
		}

		body, err := client.Get(ctx, "chat/"+guid+"/message", params)
		if err != nil {
			return fmt.Errorf("listing messages for chat %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var rawMsgs []json.RawMessage
		if err := json.Unmarshal(data, &rawMsgs); err != nil {
			return fmt.Errorf("parse messages: %w", err)
		}

		summaries := make([]MessageSummary, 0, len(rawMsgs))
		for _, r := range rawMsgs {
			summaries = append(summaries, toMessageSummary(r))
		}

		var lines []string
		for _, m := range summaries {
			lines = append(lines, formatMessageLine(m))
		}

		return printResult(cmd, summaries, lines)
	}
}

// --- chats read ---

func newChatsReadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Mark a chat as read",
		RunE:  makeRunChatsRead(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Bool("dry-run", false, "Preview without marking")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsRead(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("mark chat read", map[string]any{"guid": guid}),
				[]string{fmt.Sprintf("[dry-run] would mark chat %s as read", guid)},
			)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Post(ctx, "chat/"+guid+"/read", nil)
		if err != nil {
			return fmt.Errorf("marking chat %s as read: %w", guid, err)
		}

		_, err = ParseResponse(body)
		if err != nil {
			return err
		}

		return printResult(cmd, map[string]any{"guid": guid, "read": true}, []string{"Marked chat as read: " + guid})
	}
}

// --- chats unread ---

func newChatsUnreadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unread",
		Short: "Mark a chat as unread",
		RunE:  makeRunChatsUnread(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Bool("dry-run", false, "Preview without marking")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsUnread(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("mark chat unread", map[string]any{"guid": guid}),
				[]string{fmt.Sprintf("[dry-run] would mark chat %s as unread", guid)},
			)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Post(ctx, "chat/"+guid+"/unread", nil)
		if err != nil {
			return fmt.Errorf("marking chat %s as unread: %w", guid, err)
		}

		_, err = ParseResponse(body)
		if err != nil {
			return err
		}

		return printResult(cmd, map[string]any{"guid": guid, "unread": true}, []string{"Marked chat as unread: " + guid})
	}
}

// --- chats leave ---

func newChatsLeaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave a group chat",
		RunE:  makeRunChatsLeave(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Bool("dry-run", false, "Preview without leaving")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsLeave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("leave chat", map[string]any{"guid": guid}),
				[]string{fmt.Sprintf("[dry-run] would leave chat %s", guid)},
			)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Post(ctx, "chat/"+guid+"/leave", nil)
		if err != nil {
			return fmt.Errorf("leaving chat %s: %w", guid, err)
		}

		_, err = ParseResponse(body)
		if err != nil {
			return err
		}

		return printResult(cmd, map[string]any{"guid": guid, "left": true}, []string{"Left chat: " + guid})
	}
}

// --- chats typing ---

func newChatsTypingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "typing",
		Short: "Send or stop a typing indicator in a chat",
		RunE:  makeRunChatsTyping(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Bool("stop", false, "Stop typing indicator (default: start)")
	cmd.Flags().Bool("dry-run", false, "Preview without sending")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsTyping(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		stop, _ := cmd.Flags().GetBool("stop")

		action := "start typing"
		if stop {
			action = "stop typing"
		}

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult(action, map[string]any{"guid": guid, "stop": stop}),
				[]string{fmt.Sprintf("[dry-run] would %s in chat %s", action, guid)},
			)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var (
			body []byte
			reqErr error
		)
		if stop {
			body, reqErr = client.Delete(ctx, "chat/"+guid+"/typing")
		} else {
			body, reqErr = client.Post(ctx, "chat/"+guid+"/typing", nil)
		}
		if reqErr != nil {
			return fmt.Errorf("%s in chat %s: %w", action, guid, reqErr)
		}

		_, err = ParseResponse(body)
		if err != nil {
			return err
		}

		return printResult(cmd,
			map[string]any{"guid": guid, "typing": !stop},
			[]string{fmt.Sprintf("Typing indicator %s in chat: %s", map[bool]string{true: "stopped", false: "started"}[stop], guid)},
		)
	}
}

// --- chats count ---

func newChatsCountCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get total number of chats",
		RunE:  makeRunChatsCount(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsCount(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, "chat/count", nil)
		if err != nil {
			return fmt.Errorf("getting chat count: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("parse count: %w", err)
		}

		count := int64(0)
		if v, ok := result["total"].(float64); ok {
			count = int64(v)
		}

		return printResult(cmd, result, []string{fmt.Sprintf("Total chats: %d", count)})
	}
}

// --- chats icon (sub-group) ---

func newChatsIconCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "icon",
		Short: "Manage group chat icon",
	}
	cmd.AddCommand(newChatsIconGetCmd(factory))
	cmd.AddCommand(newChatsIconSetCmd(factory))
	cmd.AddCommand(newChatsIconRemoveCmd(factory))
	return cmd
}

// --- chats icon get ---

func newChatsIconGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get the group chat icon",
		RunE:  makeRunChatsIconGet(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().String("output", "", "File path to write image data to")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsIconGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		output, _ := cmd.Flags().GetString("output")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if output != "" {
			f, err := os.Create(output)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer f.Close()

			if err := client.Download(ctx, "chat/"+guid+"/icon", f); err != nil {
				return fmt.Errorf("downloading icon for chat %s: %w", guid, err)
			}

			result := map[string]any{"guid": guid, "output": output}
			return printResult(cmd, result, []string{fmt.Sprintf("Icon saved to: %s", output)})
		}

		// Without --output, report metadata only.
		result := map[string]any{"guid": guid, "note": "Use --output=PATH to save icon to a file"}
		return printResult(cmd, result, []string{fmt.Sprintf("Chat icon for %s: use --output=PATH to download", guid)})
	}
}

// --- chats icon set ---

func newChatsIconSetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set the group chat icon",
		RunE:  makeRunChatsIconSet(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().String("path", "", "Local path to the image file (required)")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().Bool("dry-run", false, "Preview without uploading")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsIconSet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		path, _ := cmd.Flags().GetString("path")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("set chat icon", map[string]any{"guid": guid, "path": path}),
				[]string{fmt.Sprintf("[dry-run] would set icon for chat %s from %s", guid, path)},
			)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"path": path,
		}

		body, err := client.Post(ctx, "chat/"+guid+"/icon", reqBody)
		if err != nil {
			return fmt.Errorf("setting icon for chat %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var result map[string]any
		_ = json.Unmarshal(data, &result)
		if result == nil {
			result = map[string]any{"guid": guid, "updated": true}
		}

		return printResult(cmd, result, []string{fmt.Sprintf("Icon set for chat: %s", guid)})
	}
}

// --- chats icon remove ---

func newChatsIconRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove the group chat icon",
		RunE:  makeRunChatsIconRemove(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	cmd.Flags().Bool("dry-run", false, "Preview without removing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunChatsIconRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("remove chat icon", map[string]any{"guid": guid}),
				[]string{fmt.Sprintf("[dry-run] would remove icon for chat %s", guid)},
			)
		}

		if err := confirmDestructive(cmd, "remove chat icon"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Delete(ctx, "chat/"+guid+"/icon")
		if err != nil {
			return fmt.Errorf("removing icon for chat %s: %w", guid, err)
		}

		_, err = ParseResponse(body)
		if err != nil {
			return err
		}

		return printResult(cmd,
			map[string]any{"guid": guid, "icon_removed": true},
			[]string{"Icon removed for chat: " + guid},
		)
	}
}
