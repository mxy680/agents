package imessage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// ChatSummary is a condensed chat representation.
type ChatSummary struct {
	GUID         string   `json:"guid"`
	DisplayName  string   `json:"display_name,omitempty"`
	ChatID       string   `json:"chat_identifier"`
	IsArchived   bool     `json:"is_archived"`
	Participants []string `json:"participants,omitempty"`
	LastMessage  string   `json:"last_message,omitempty"`
}

// MessageSummary is a condensed message representation.
type MessageSummary struct {
	GUID          string `json:"guid"`
	Text          string `json:"text,omitempty"`
	IsFromMe      bool   `json:"is_from_me"`
	DateCreated   int64  `json:"date_created"`
	Handle        string `json:"handle,omitempty"`
	ChatGUID      string `json:"chat_guid,omitempty"`
	Subject       string `json:"subject,omitempty"`
	HasAttachment bool   `json:"has_attachments"`
}

// AttachmentSummary is a condensed attachment representation.
type AttachmentSummary struct {
	GUID         string `json:"guid"`
	FileName     string `json:"transfer_name,omitempty"`
	MIMEType     string `json:"mime_type,omitempty"`
	TotalBytes   int64  `json:"total_bytes"`
	IsOutgoing   bool   `json:"is_outgoing"`
	DateCreated  int64  `json:"created_date"`
}

// HandleSummary is a condensed handle representation.
type HandleSummary struct {
	Address     string `json:"address"`
	Service     string `json:"service"`
	Country     string `json:"country,omitempty"`
	UncanonID   string `json:"uncanonicalized_id,omitempty"`
}

// ContactSummary is a condensed contact representation.
type ContactSummary struct {
	ID          string `json:"id,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Phones      []string `json:"phones,omitempty"`
	Emails      []string `json:"emails,omitempty"`
}

// ScheduledMessageSummary is a condensed scheduled message representation.
type ScheduledMessageSummary struct {
	ID       int    `json:"id"`
	ChatGUID string `json:"chat_guid"`
	Message  string `json:"message"`
	SendDate int64  `json:"scheduled_for"`
	Status   string `json:"status,omitempty"`
}

// ServerInfo holds BlueBubbles server information.
type ServerInfo struct {
	OSVersion     string `json:"os_version,omitempty"`
	ServerVersion string `json:"server_version,omitempty"`
	MacModel      string `json:"detected_icloud,omitempty"`
	PrivateAPI    bool   `json:"private_api"`
	ProxyService  string `json:"proxy_service,omitempty"`
}

// WebhookSummary is a condensed webhook representation.
type WebhookSummary struct {
	ID     int    `json:"id"`
	URL    string `json:"url"`
	Events []string `json:"events,omitempty"`
}

// FindMyDevice is a device from FindMy.
type FindMyDevice struct {
	ID       string  `json:"id"`
	Name     string  `json:"name,omitempty"`
	Battery  float64 `json:"batteryLevel,omitempty"`
	Location any     `json:"location,omitempty"`
}

// FindMyFriend is a friend from FindMy.
type FindMyFriend struct {
	ID       string `json:"id"`
	Handle   string `json:"handle,omitempty"`
	Name     string `json:"firstName,omitempty"`
	Location any    `json:"location,omitempty"`
}

// truncate trims s to maxLen characters, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// confirmDestructive checks if --confirm is set or prompts implicitly (for CLI use,
// we require --confirm on destructive ops without a tty).
func confirmDestructive(cmd *cobra.Command, action string) error {
	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		return fmt.Errorf("destructive action %q requires --confirm flag", action)
	}
	return nil
}

// dryRunResult returns a dry-run placeholder result.
func dryRunResult(action string, details map[string]any) map[string]any {
	result := map[string]any{
		"dry_run": true,
		"action":  action,
	}
	for k, v := range details {
		result[k] = v
	}
	return result
}

// formatTimestamp converts a BlueBubbles timestamp (ms since epoch or cocoa epoch)
// to a human-readable string.
func formatTimestamp(ts int64) string {
	if ts <= 0 {
		return ""
	}
	// BlueBubbles uses milliseconds since Unix epoch.
	t := time.Unix(ts/1000, (ts%1000)*int64(time.Millisecond))
	return t.Format("2006-01-02 15:04:05")
}

// toChatSummary converts a raw JSON chat object to a ChatSummary.
func toChatSummary(raw json.RawMessage) ChatSummary {
	var m map[string]any
	json.Unmarshal(raw, &m)

	summary := ChatSummary{
		GUID:       getString(m, "guid"),
		DisplayName: getString(m, "displayName"),
		ChatID:     getString(m, "chatIdentifier"),
		IsArchived: getBool(m, "isArchived"),
	}

	if parts, ok := m["participants"].([]any); ok {
		for _, p := range parts {
			if pm, ok := p.(map[string]any); ok {
				if h, ok := pm["handle"].(map[string]any); ok {
					summary.Participants = append(summary.Participants, getString(h, "address"))
				}
			}
		}
	}

	if lm, ok := m["lastMessage"].(map[string]any); ok {
		summary.LastMessage = truncate(getString(lm, "text"), 80)
	}

	return summary
}

// toMessageSummary converts a raw JSON message object to a MessageSummary.
func toMessageSummary(raw json.RawMessage) MessageSummary {
	var m map[string]any
	json.Unmarshal(raw, &m)

	summary := MessageSummary{
		GUID:        getString(m, "guid"),
		Text:        getString(m, "text"),
		IsFromMe:    getBool(m, "isFromMe"),
		DateCreated: getInt64(m, "dateCreated"),
		Subject:     getString(m, "subject"),
	}

	if h, ok := m["handle"].(map[string]any); ok {
		summary.Handle = getString(h, "address")
	}

	if chats, ok := m["chats"].([]any); ok && len(chats) > 0 {
		if cm, ok := chats[0].(map[string]any); ok {
			summary.ChatGUID = getString(cm, "guid")
		}
	}

	if attachments, ok := m["attachments"].([]any); ok {
		summary.HasAttachment = len(attachments) > 0
	}

	return summary
}

// toAttachmentSummary converts a raw JSON attachment object to an AttachmentSummary.
func toAttachmentSummary(raw json.RawMessage) AttachmentSummary {
	var m map[string]any
	json.Unmarshal(raw, &m)
	return AttachmentSummary{
		GUID:        getString(m, "guid"),
		FileName:    getString(m, "transferName"),
		MIMEType:    getString(m, "mimeType"),
		TotalBytes:  getInt64(m, "totalBytes"),
		IsOutgoing:  getBool(m, "isOutgoing"),
		DateCreated: getInt64(m, "createdDate"),
	}
}

// toHandleSummary converts a raw JSON handle object to a HandleSummary.
func toHandleSummary(raw json.RawMessage) HandleSummary {
	var m map[string]any
	json.Unmarshal(raw, &m)
	return HandleSummary{
		Address:   getString(m, "address"),
		Service:   getString(m, "service"),
		Country:   getString(m, "country"),
		UncanonID: getString(m, "uncanonicalizedId"),
	}
}

// toContactSummary converts a raw JSON contact object to a ContactSummary.
func toContactSummary(raw json.RawMessage) ContactSummary {
	var m map[string]any
	json.Unmarshal(raw, &m)

	summary := ContactSummary{
		ID:          getString(m, "id"),
		FirstName:   getString(m, "firstName"),
		LastName:    getString(m, "lastName"),
		DisplayName: getString(m, "displayName"),
	}

	if phones, ok := m["phoneNumbers"].([]any); ok {
		for _, p := range phones {
			if pm, ok := p.(map[string]any); ok {
				summary.Phones = append(summary.Phones, getString(pm, "address"))
			}
		}
	}
	if emails, ok := m["emails"].([]any); ok {
		for _, e := range emails {
			if em, ok := e.(map[string]any); ok {
				summary.Emails = append(summary.Emails, getString(em, "address"))
			}
		}
	}

	return summary
}

// formatChatLine formats a single chat for text output.
func formatChatLine(c ChatSummary) string {
	name := c.DisplayName
	if name == "" {
		name = c.ChatID
	}
	line := fmt.Sprintf("%-40s  %s", truncate(name, 38), c.GUID)
	if c.LastMessage != "" {
		line += fmt.Sprintf("\n  Last: %s", c.LastMessage)
	}
	return line
}

// formatMessageLine formats a single message for text output.
func formatMessageLine(m MessageSummary) string {
	direction := "<-"
	if m.IsFromMe {
		direction = "->"
	}
	ts := formatTimestamp(m.DateCreated)
	handle := m.Handle
	if handle == "" {
		handle = "me"
	}
	return fmt.Sprintf("[%s] %s %s: %s", ts, direction, handle, truncate(m.Text, 80))
}

// printResult is a convenience wrapper for cli.PrintResult.
func printResult(cmd *cobra.Command, v any, textLines []string) error {
	return cli.PrintResult(cmd, v, textLines)
}

// getString safely extracts a string from a map.
func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getBool safely extracts a bool from a map.
func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// getInt64 safely extracts an int64 from a map (JSON numbers are float64).
func getInt64(m map[string]any, key string) int64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return int64(f)
		}
	}
	return 0
}
