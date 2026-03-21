package x

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newBlocksCmd builds the "blocks" subcommand group.
func newBlocksCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "blocks",
		Short:   "Manage blocked users",
		Aliases: []string{"block"},
	}
	cmd.AddCommand(newBlocksBlockCmd(factory))
	cmd.AddCommand(newBlocksUnblockCmd(factory))
	return cmd
}

// newBlocksBlockCmd builds the "blocks block" command.
func newBlocksBlockCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Block a user",
		RunE:  makeRunBlocksBlock(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to block (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without blocking")
	return cmd
}

// newBlocksUnblockCmd builds the "blocks unblock" command.
func newBlocksUnblockCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unblock",
		Short: "Unblock a user",
		RunE:  makeRunBlocksUnblock(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to unblock (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without unblocking")
	return cmd
}

// --- RunE implementations ---

func makeRunBlocksBlock(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("block user %s", userID), map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_id", userID)

		resp, err := client.Post(ctx, "/i/api/1.1/blocks/create.json", body)
		if err != nil {
			return fmt.Errorf("blocking user %s: %w", userID, err)
		}

		var result json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("blocking user %s: %w", userID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "blocked", "user_id": userID})
		}
		fmt.Printf("Blocked user: %s\n", userID)
		return nil
	}
}

func makeRunBlocksUnblock(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("unblock user %s", userID), map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_id", userID)

		resp, err := client.Post(ctx, "/i/api/1.1/blocks/destroy.json", body)
		if err != nil {
			return fmt.Errorf("unblocking user %s: %w", userID, err)
		}

		var result json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("unblocking user %s: %w", userID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unblocked", "user_id": userID})
		}
		fmt.Printf("Unblocked user: %s\n", userID)
		return nil
	}
}
