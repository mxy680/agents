package canvas

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newNotificationsCmd returns the parent "notifications" command with all subcommands attached.
func newNotificationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notifications",
		Short:   "Manage Canvas notifications and communication preferences",
		Aliases: []string{"notif"},
	}

	cmd.AddCommand(newNotificationsListCmd(factory))
	cmd.AddCommand(newNotificationsPreferencesCmd(factory))
	cmd.AddCommand(newNotificationsUpdatePreferenceCmd(factory))

	return cmd
}

func newNotificationsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List account notifications for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self/account_notifications", nil)
			if err != nil {
				return err
			}

			var notifications []AccountNotificationSummary
			if err := json.Unmarshal(data, &notifications); err != nil {
				return fmt.Errorf("parse notifications: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(notifications)
			}

			if len(notifications) == 0 {
				fmt.Println("No notifications found.")
				return nil
			}
			for _, n := range notifications {
				active := ""
				if n.StartAt != "" || n.EndAt != "" {
					active = fmt.Sprintf(" [%s → %s]", n.StartAt, n.EndAt)
				}
				fmt.Printf("%-6d  %s%s\n", n.ID, truncate(n.Subject, 60), active)
			}
			return nil
		},
	}

	return cmd
}

func newNotificationsPreferencesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preferences",
		Short: "Get notification preferences for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/users/self/communication_channels/email/self/notification_preferences"
			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse notification preferences: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}

			prefs, _ := result["notification_preferences"].([]any)
			if len(prefs) == 0 {
				fmt.Println("No notification preferences found.")
				return nil
			}
			for _, p := range prefs {
				pref, ok := p.(map[string]any)
				if !ok {
					continue
				}
				category, _ := pref["notification"].(string)
				frequency, _ := pref["frequency"].(string)
				fmt.Printf("%-40s  %s\n", truncate(category, 40), frequency)
			}
			return nil
		},
	}

	return cmd
}

func newNotificationsUpdatePreferenceCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-preference",
		Short: "Update a notification preference frequency",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			category, _ := cmd.Flags().GetString("category")
			frequency, _ := cmd.Flags().GetString("frequency")
			if category == "" {
				return fmt.Errorf("--category is required")
			}
			if frequency == "" {
				frequency = "immediately"
			}

			validFrequencies := map[string]bool{
				"immediately": true, "daily": true, "weekly": true, "never": true,
			}
			if !validFrequencies[frequency] {
				return fmt.Errorf("invalid frequency %q: must be immediately, daily, weekly, or never", frequency)
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "update notification preference: "+category, map[string]any{
					"category": category, "frequency": frequency,
				})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/users/self/communication_channels/email/self/notification_preferences/" + category
			body := map[string]any{
				"notification_preferences": []map[string]any{
					{"notification": category, "frequency": frequency},
				},
			}
			data, err := client.Put(ctx, path, body)
			if err != nil {
				return err
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse preference response: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}
			fmt.Printf("Notification preference '%s' updated to '%s'\n", category, frequency)
			return nil
		},
	}

	cmd.Flags().String("category", "", "Notification category (required)")
	cmd.Flags().String("frequency", "immediately", "Frequency: immediately|daily|weekly|never")
	return cmd
}
