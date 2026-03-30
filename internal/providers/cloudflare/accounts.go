package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAccountsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloudflare accounts",
		RunE:  makeRunAccountsList(factory),
	}
	cmd.Flags().Int("per-page", 50, "Number of accounts per page")
	return cmd
}

func makeRunAccountsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		perPage, _ := cmd.Flags().GetInt("per-page")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/accounts?per_page=%d", perPage), nil, &resp); err != nil {
			return fmt.Errorf("listing accounts: %w", err)
		}

		accounts := make([]AccountSummary, 0, len(resp))
		for _, a := range resp {
			accounts = append(accounts, toAccountSummary(a))
		}

		return printAccounts(cmd, accounts)
	}
}

func newAccountsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get account details",
		RunE:  makeRunAccountsGet(factory),
	}
	cmd.Flags().String("account", "", "Account ID (required)")
	_ = cmd.MarkFlagRequired("account")
	return cmd
}

func makeRunAccountsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		accountID, _ := cmd.Flags().GetString("account")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s", accountID), nil, &data); err != nil {
			return fmt.Errorf("getting account %q: %w", accountID, err)
		}

		account := toAccountSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(account)
		}

		lines := []string{
			fmt.Sprintf("ID:   %s", account.ID),
			fmt.Sprintf("Name: %s", account.Name),
			fmt.Sprintf("Type: %s", account.Type),
		}
		cli.PrintText(lines)
		return nil
	}
}
