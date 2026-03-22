package x

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// NotificationSummary is a condensed representation of an X notification.
type NotificationSummary struct {
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// newNotificationsCmd builds the "notifications" subcommand group.
func newNotificationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notifications",
		Short:   "View X notifications",
		Aliases: []string{"notif"},
	}
	cmd.AddCommand(newNotificationsAllCmd(factory))
	cmd.AddCommand(newNotificationsMentionsCmd(factory))
	cmd.AddCommand(newNotificationsVerifiedCmd(factory))
	return cmd
}

func newNotificationsAllCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Get all notifications",
		RunE:  makeRunNotifications(factory, "/i/api/2/notifications/all.json"),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of notifications")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func newNotificationsMentionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mentions",
		Short: "Get mention notifications",
		RunE:  makeRunNotifications(factory, "/i/api/2/notifications/mentions.json"),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of notifications")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func newNotificationsVerifiedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verified",
		Short: "Get verified account notifications",
		RunE:  makeRunNotifications(factory, "/i/api/2/notifications/verified.json"),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of notifications")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// makeRunNotifications returns a RunE func for any notifications endpoint.
func makeRunNotifications(factory ClientFactory, path string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", fmt.Sprintf("%d", limit))
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("fetching notifications: %w", err)
		}

		var raw json.RawMessage
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decode notifications response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(raw)
		}

		// For text output, extract notification entries from the timeline.
		notifications := extractNotifications(raw)
		if len(notifications) == 0 {
			fmt.Println("No notifications found.")
			return nil
		}

		lines := make([]string, 0, len(notifications)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-60s", "ID", "TYPE", "MESSAGE"))
		for _, n := range notifications {
			lines = append(lines, fmt.Sprintf("%-20s  %-15s  %-60s",
				truncate(n.ID, 20),
				truncate(n.Type, 15),
				truncate(n.Message, 60),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

// extractNotifications parses notification entries from the v2 notifications API response.
func extractNotifications(data json.RawMessage) []NotificationSummary {
	var envelope struct {
		Timeline struct {
			Instructions []struct {
				AddEntries *struct {
					Entries []json.RawMessage `json:"entries"`
				} `json:"addEntries"`
			} `json:"instructions"`
		} `json:"timeline"`
		GlobalObjects struct {
			Notifications map[string]json.RawMessage `json:"notifications"`
		} `json:"globalObjects"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil
	}

	var notifications []NotificationSummary

	// Use globalObjects.notifications if available.
	for id, raw := range envelope.GlobalObjects.Notifications {
		var notif struct {
			Type    string `json:"type"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			TimestampMs string `json:"timestampMs"`
		}
		_ = json.Unmarshal(raw, &notif)
		notifications = append(notifications, NotificationSummary{
			ID:        id,
			Type:      notif.Type,
			Message:   notif.Message.Text,
			Timestamp: notif.TimestampMs,
		})
	}

	if len(notifications) > 0 {
		return notifications
	}

	// Fall back to timeline entries.
	for _, instr := range envelope.Timeline.Instructions {
		if instr.AddEntries == nil {
			continue
		}
		for _, entryRaw := range instr.AddEntries.Entries {
			var entry struct {
				EntryID string `json:"entryId"`
				Content struct {
					TimelineModule *struct {
						Items []struct {
							Item struct {
								Content struct {
									Notification *struct {
										Message struct {
											Text string `json:"text"`
										} `json:"message"`
									} `json:"notification"`
								} `json:"content"`
							} `json:"item"`
						} `json:"items"`
					} `json:"timelineModule"`
				} `json:"content"`
			}
			if err := json.Unmarshal(entryRaw, &entry); err != nil {
				continue
			}
			if entry.Content.TimelineModule != nil {
				for _, item := range entry.Content.TimelineModule.Items {
					if item.Item.Content.Notification != nil {
						notifications = append(notifications, NotificationSummary{
							ID:      entry.EntryID,
							Message: item.Item.Content.Notification.Message.Text,
						})
					}
				}
			} else {
				// Simple entry — include raw ID.
				notifications = append(notifications, NotificationSummary{
					ID: entry.EntryID,
				})
			}
		}
	}

	return notifications
}
