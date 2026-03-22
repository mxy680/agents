package imessage

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newICloudCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "icloud",
		Short: "iCloud account information",
	}

	cmd.AddCommand(newICloudAccountCmd(factory))
	cmd.AddCommand(newICloudChangeAliasCmd(factory))
	cmd.AddCommand(newICloudContactCardCmd(factory))

	return cmd
}

func newICloudAccountCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Get iCloud account information",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunICloudAccount(factory)
	return cmd
}

func makeRunICloudAccount(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "icloud/account", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return printResult(cmd, data, []string{string(data)})
		}

		lines := []string{
			fmt.Sprintf("Apple ID: %s", getString(m, "appleId")),
			fmt.Sprintf("Name:     %s", getString(m, "fullName")),
		}
		return printResult(cmd, m, lines)
	}
}

func newICloudChangeAliasCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-alias",
		Short: "Change iCloud account alias",
	}
	cmd.Flags().String("alias", "", "New alias to set (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("alias")
	cmd.RunE = makeRunICloudChangeAlias(factory)
	return cmd
}

func makeRunICloudChangeAlias(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		alias, _ := cmd.Flags().GetString("alias")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("icloud change-alias", map[string]any{"alias": alias})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would change alias to: %s", alias)})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), "icloud/account/alias", map[string]any{"alias": alias})
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{fmt.Sprintf("Alias changed to: %s", alias)})
	}
}

func newICloudContactCardCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contact-card",
		Short: "Get iCloud contact card",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunICloudContactCard(factory)
	return cmd
}

func makeRunICloudContactCard(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "icloud/contact", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return printResult(cmd, data, []string{string(data)})
		}

		lines := []string{
			fmt.Sprintf("Name:  %s %s", getString(m, "firstName"), getString(m, "lastName")),
			fmt.Sprintf("Email: %s", getString(m, "email")),
			fmt.Sprintf("Phone: %s", getString(m, "phoneNumber")),
		}
		return printResult(cmd, m, lines)
	}
}
