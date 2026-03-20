package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSecretsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secrets",
		Aliases: []string{"secret"},
		Short:   "Manage Edge Function secrets",
	}
	cmd.AddCommand(
		newSecretsListCmd(factory),
		newSecretsCreateCmd(factory),
		newSecretsDeleteCmd(factory),
	)
	return cmd
}

// printSecretSummaries outputs secret summaries as JSON or a formatted text table.
// Secret values are masked in table output for security.
func printSecretSummaries(cmd *cobra.Command, secrets []SecretSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(secrets)
	}
	if len(secrets) == 0 {
		fmt.Println("No secrets found.")
		return nil
	}
	lines := make([]string, 0, len(secrets)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %s", "NAME", "VALUE"))
	for _, s := range secrets {
		lines = append(lines, fmt.Sprintf("%-40s  %s", truncate(s.Name, 40), maskKey(s.Value)))
	}
	cli.PrintText(lines)
	return nil
}

// --- list ---

func newSecretsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Edge Function secrets for a project",
		RunE:  makeRunSecretsList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunSecretsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/secrets", ref), nil)
		if err != nil {
			return fmt.Errorf("listing secrets: %w", err)
		}

		var raw []map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		secrets := make([]SecretSummary, 0, len(raw))
		for _, r := range raw {
			name, _ := r["name"].(string)
			value, _ := r["value"].(string)
			secrets = append(secrets, SecretSummary{Name: name, Value: value})
		}
		return printSecretSummaries(cmd, secrets)
	}
}

// --- create ---

func newSecretsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create or update an Edge Function secret",
		RunE:  makeRunSecretsCreate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("name", "", "Secret name (required)")
	cmd.Flags().String("value", "", "Secret value (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func makeRunSecretsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		name, _ := cmd.Flags().GetString("name")
		value, _ := cmd.Flags().GetString("value")

		if dryRunResult(cmd, fmt.Sprintf("Would create secret %q in project %s", name, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// API expects an array of secrets
		bodySlice := []map[string]string{{"name": name, "value": value}}
		bodyBytes, err := json.Marshal(bodySlice)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		if _, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/secrets", ref), bytes.NewReader(bodyBytes)); err != nil {
			return fmt.Errorf("creating secret: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "created", "name": name})
		}
		fmt.Printf("Created secret: %s\n", name)
		return nil
	}
}

// --- delete ---

func newSecretsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an Edge Function secret (irreversible)",
		RunE:  makeRunSecretsDelete(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("name", "", "Secret name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunSecretsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		name, _ := cmd.Flags().GetString("name")

		if dryRunResult(cmd, fmt.Sprintf("Would permanently delete secret %q from project %s", name, ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting secret %q is irreversible", name)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// API expects an array of secret names
		bodySlice := []string{name}
		bodyBytes, err := json.Marshal(bodySlice)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/secrets", ref), bytes.NewReader(bodyBytes)); err != nil {
			return fmt.Errorf("deleting secret %q: %w", name, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "name": name})
		}
		fmt.Printf("Deleted secret: %s\n", name)
		return nil
	}
}
