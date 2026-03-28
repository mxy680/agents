package linkedin

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newNotificationsCmd builds the "notifications" (alias: notif) subcommand group.
func newNotificationsCmd(factory ClientFactory) *cobra.Command {
	notifCmd := &cobra.Command{
		Use:     "notifications",
		Short:   "Manage your LinkedIn notifications",
		Aliases: []string{"notif"},
	}
	notifCmd.AddCommand(newNotificationsListCmd(factory))
	notifCmd.AddCommand(newNotificationsMarkReadCmd(factory))
	return notifCmd
}

// newNotificationsListCmd builds the "notifications list" command.
func newNotificationsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your LinkedIn notifications",
		Long:  "List recent LinkedIn notifications including profile views, reactions, and more.",
		RunE:  makeRunNotificationsList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of notifications to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

func makeRunNotificationsList(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

// newNotificationsMarkReadCmd builds the "notifications mark-read" command.
func newNotificationsMarkReadCmd(_ ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-read",
		Short: "Mark all notifications as read",
		RunE:  makeRunNotificationsMarkRead(),
	}
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	return cmd
}

func makeRunNotificationsMarkRead() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "mark all notifications as read", nil)
		}
		return errEndpointDeprecated
	}
}

// printNotificationSummaries outputs notification summaries as JSON or text.
func printNotificationSummaries(cmd *cobra.Command, notifications []NotificationSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(notifications)
	}
	if len(notifications) == 0 {
		fmt.Println("No notifications found.")
		return nil
	}
	lines := make([]string, 0, len(notifications)+1)
	lines = append(lines, fmt.Sprintf("%-45s  %-50s  %-16s  %-5s", "ID", "TEXT", "DATE", "READ"))
	for _, n := range notifications {
		readStr := "false"
		if n.IsRead {
			readStr = "true"
		}
		lines = append(lines, fmt.Sprintf("%-45s  %-50s  %-16s  %-5s",
			truncate(n.ID, 45),
			truncate(n.Text, 50),
			formatTimestamp(n.Timestamp),
			readStr,
		))
	}
	cli.PrintText(lines)
	return nil
}
