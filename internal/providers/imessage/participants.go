package imessage

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newParticipantsCmd returns the parent "participants" command with all subcommands attached.
func newParticipantsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "participants",
		Short:   "Manage group chat participants",
		Aliases: []string{"participant"},
	}

	cmd.AddCommand(newParticipantsAddCmd(factory))
	cmd.AddCommand(newParticipantsRemoveCmd(factory))

	return cmd
}

// --- participants add ---

func newParticipantsAddCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a participant to a group chat",
		RunE:  makeRunParticipantsAdd(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().String("address", "", "Phone number or email to add (required)")
	_ = cmd.MarkFlagRequired("address")
	cmd.Flags().Bool("dry-run", false, "Preview without adding")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunParticipantsAdd(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		address, _ := cmd.Flags().GetString("address")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("add participant", map[string]any{"guid": guid, "address": address}),
				[]string{fmt.Sprintf("[dry-run] would add %s to chat %s", address, guid)},
			)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"address": address,
		}

		body, err := client.Post(ctx, "chat/"+guid+"/participant", reqBody)
		if err != nil {
			return fmt.Errorf("adding participant %s to chat %s: %w", address, guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toChatSummary(data)
		lines := []string{fmt.Sprintf("Added %s to chat %s", address, guid)}
		if len(summary.Participants) > 0 {
			lines = append(lines, "  Current participants:")
			for _, p := range summary.Participants {
				lines = append(lines, "    - "+p)
			}
		}

		return printResult(cmd, summary, lines)
	}
}

// --- participants remove ---

func newParticipantsRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a participant from a group chat",
		RunE:  makeRunParticipantsRemove(factory),
	}
	cmd.Flags().String("guid", "", "Chat GUID (required)")
	_ = cmd.MarkFlagRequired("guid")
	cmd.Flags().String("address", "", "Phone number or email to remove (required)")
	_ = cmd.MarkFlagRequired("address")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	cmd.Flags().Bool("dry-run", false, "Preview without removing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunParticipantsRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		address, _ := cmd.Flags().GetString("address")

		if cli.IsDryRun(cmd) {
			return printResult(cmd,
				dryRunResult("remove participant", map[string]any{"guid": guid, "address": address}),
				[]string{fmt.Sprintf("[dry-run] would remove %s from chat %s", address, guid)},
			)
		}

		if err := confirmDestructive(cmd, "remove participant"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"address": address,
		}

		body, err := client.Post(ctx, "chat/"+guid+"/participant/remove", reqBody)
		if err != nil {
			return fmt.Errorf("removing participant %s from chat %s: %w", address, guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toChatSummary(data)
		lines := []string{fmt.Sprintf("Removed %s from chat %s", address, guid)}
		if len(summary.Participants) > 0 {
			lines = append(lines, "  Remaining participants:")
			for _, p := range summary.Participants {
				lines = append(lines, "    - "+p)
			}
		}

		return printResult(cmd, summary, lines)
	}
}
