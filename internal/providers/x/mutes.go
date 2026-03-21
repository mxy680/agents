package x

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newMutesCmd builds the "mutes" subcommand group.
func newMutesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mutes",
		Short:   "Manage muted users",
		Aliases: []string{"mute"},
	}
	cmd.AddCommand(newMutesMuteCmd(factory))
	cmd.AddCommand(newMutesUnmuteCmd(factory))
	return cmd
}

// newMutesMuteCmd builds the "mutes mute" command.
func newMutesMuteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mute",
		Short: "Mute a user",
		RunE:  makeRunMutesMute(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to mute (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without muting")
	return cmd
}

// newMutesUnmuteCmd builds the "mutes unmute" command.
func newMutesUnmuteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmute",
		Short: "Unmute a user",
		RunE:  makeRunMutesUnmute(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to unmute (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without unmuting")
	return cmd
}

// --- RunE implementations ---

func makeRunMutesMute(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("mute user %s", userID), map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_id", userID)

		resp, err := client.Post(ctx, "/i/api/1.1/mutes/users/create.json", body)
		if err != nil {
			return fmt.Errorf("muting user %s: %w", userID, err)
		}

		var result json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("muting user %s: %w", userID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "muted", "user_id": userID})
		}
		fmt.Printf("Muted user: %s\n", userID)
		return nil
	}
}

func makeRunMutesUnmute(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("unmute user %s", userID), map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_id", userID)

		resp, err := client.Post(ctx, "/i/api/1.1/mutes/users/destroy.json", body)
		if err != nil {
			return fmt.Errorf("unmuting user %s: %w", userID, err)
		}

		var result json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("unmuting user %s: %w", userID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unmuted", "user_id": userID})
		}
		fmt.Printf("Unmuted user: %s\n", userID)
		return nil
	}
}
