package imessage

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newMacCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mac",
		Short: "macOS system commands",
	}

	cmd.AddCommand(newMacLockCmd(factory))
	cmd.AddCommand(newMacRestartMessagesCmd(factory))

	return cmd
}

func newMacLockCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock the Mac",
	}
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunMacLock(factory)
	return cmd
}

func makeRunMacLock(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("mac lock", nil)
			return printResult(cmd, result, []string{"[dry-run] Would lock the Mac."})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), "mac/lock", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{"Mac locked."})
	}
}

func newMacRestartMessagesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart-messages",
		Short: "Restart the Messages app on Mac",
	}
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunMacRestartMessages(factory)
	return cmd
}

func makeRunMacRestartMessages(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("mac restart-messages", nil)
			return printResult(cmd, result, []string{"[dry-run] Would restart the Messages app."})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), "mac/imessage/restart", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{fmt.Sprintf("Messages app restarted: %s", string(data))})
	}
}
