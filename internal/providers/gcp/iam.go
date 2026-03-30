package gcp

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newIAMListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts for a project",
		RunE:  makeRunIAMList(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	return cmd
}

func makeRunIAMList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/serviceAccounts", client.iamURL, project)
		var resp struct {
			Accounts []map[string]any `json:"accounts"`
		}
		if err := client.doJSON(ctx, http.MethodGet, url, nil, &resp); err != nil {
			return fmt.Errorf("listing service accounts: %w", err)
		}

		summaries := make([]ServiceAccountSummary, 0, len(resp.Accounts))
		for _, a := range resp.Accounts {
			summaries = append(summaries, toServiceAccountSummary(a))
		}

		return printServiceAccountSummaries(cmd, summaries)
	}
}

func newIAMCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service account",
		RunE:  makeRunIAMCreate(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("account-id", "", "Service account ID (required, e.g. my-service-account)")
	cmd.Flags().String("display-name", "", "Display name for the service account")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("account-id")
	return cmd
}

func makeRunIAMCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		accountID, _ := cmd.Flags().GetString("account-id")
		displayName, _ := cmd.Flags().GetString("display-name")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if displayName == "" {
			displayName = accountID
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create service account %q on project %q", accountID, project), map[string]any{
				"action":      "create",
				"project":     project,
				"accountId":   accountID,
				"displayName": displayName,
			})
		}

		body := map[string]any{
			"accountId": accountID,
			"serviceAccount": map[string]any{
				"displayName": displayName,
			},
		}

		url := fmt.Sprintf("%s/projects/%s/serviceAccounts", client.iamURL, project)
		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, url, body, &data); err != nil {
			return fmt.Errorf("creating service account %q: %w", accountID, err)
		}

		account := toServiceAccountSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(account)
		}
		fmt.Printf("Created service account: %s\n", account.Email)
		return nil
	}
}

func newIAMCreateKeyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-key",
		Short: "Create a JSON key for a service account (returns base64-encoded key file)",
		RunE:  makeRunIAMCreateKey(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("email", "", "Service account email (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunIAMCreateKey(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		email, _ := cmd.Flags().GetString("email")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create key for service account %q", email), map[string]any{
				"action":  "create-key",
				"project": project,
				"email":   email,
			})
		}

		url := fmt.Sprintf("%s/projects/%s/serviceAccounts/%s/keys", client.iamURL, project, email)
		var data map[string]any
		// Empty body uses defaults: JSON key type, no key algorithm override.
		if err := client.doJSON(ctx, http.MethodPost, url, map[string]any{}, &data); err != nil {
			return fmt.Errorf("creating key for service account %q: %w", email, err)
		}

		key := toServiceAccountKey(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(key)
		}
		lines := []string{
			fmt.Sprintf("Name:             %s", key.Name),
			fmt.Sprintf("Key Type:         %s", key.KeyType),
			fmt.Sprintf("Valid After:      %s", key.ValidAfterTime),
			fmt.Sprintf("Private Key Data: %s", key.PrivateKeyData),
		}
		cli.PrintText(lines)
		warnf("Store the private key data securely — it cannot be retrieved again.")
		return nil
	}
}

func newIAMDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a service account (irreversible)",
		RunE:  makeRunIAMDelete(factory),
	}
	cmd.Flags().String("project", "", "GCP project ID (falls back to GCP_PROJECT_ID)")
	cmd.Flags().String("email", "", "Service account email (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without making changes")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunIAMDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		flagProject, _ := cmd.Flags().GetString("project")
		email, _ := cmd.Flags().GetString("email")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		project, err := client.resolveProject(flagProject)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete service account %q", email), map[string]any{
				"action":  "delete",
				"project": project,
				"email":   email,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		url := fmt.Sprintf("%s/projects/%s/serviceAccounts/%s", client.iamURL, project, email)
		if _, err := client.do(ctx, http.MethodDelete, url, nil); err != nil {
			return fmt.Errorf("deleting service account %q: %w", email, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "email": email})
		}
		fmt.Printf("Deleted service account: %s\n", email)
		return nil
	}
}
