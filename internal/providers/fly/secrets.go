package fly

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

const (
	secretsListQuery = `query($appName: String!) {
  app(name: $appName) {
    secrets {
      name
      digest
      createdAt
    }
  }
}`

	secretsSetMutation = `mutation($appId: ID!, $secrets: [SecretInput!]!) {
  setSecrets(input: {appId: $appId, secrets: $secrets}) {
    app {
      secrets {
        name
        digest
        createdAt
      }
    }
  }
}`

	secretsUnsetMutation = `mutation($appId: ID!, $keys: [String!]!) {
  unsetSecrets(input: {appId: $appId, keys: $keys}) {
    app {
      secrets {
        name
        digest
        createdAt
      }
    }
  }
}`
)

func newSecretsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets for an app",
		RunE:  makeRunSecretsList(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	_ = cmd.MarkFlagRequired("app")
	return cmd
}

func makeRunSecretsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var result struct {
			App struct {
				Secrets []SecretSummary `json:"secrets"`
			} `json:"app"`
		}
		if err := client.graphQL(ctx, secretsListQuery, map[string]any{"appName": app}, &result); err != nil {
			return fmt.Errorf("listing secrets for app %q: %w", app, err)
		}

		return printSecretSummaries(cmd, result.App.Secrets)
	}
}

func newSecretsSetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a secret for an app",
		RunE:  makeRunSecretsSet(factory),
	}
	cmd.Flags().String("app", "", "App name / app ID (required)")
	cmd.Flags().String("key", "", "Secret key name (required)")
	cmd.Flags().String("value", "", "Secret value (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("key")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func makeRunSecretsSet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would set secret %q in app %q", key, app), map[string]any{
				"action": "set",
				"app":    app,
				"key":    key,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		variables := map[string]any{
			"appId": app,
			"secrets": []map[string]any{
				{"key": key, "value": value},
			},
		}

		var result struct {
			SetSecrets struct {
				App struct {
					Secrets []SecretSummary `json:"secrets"`
				} `json:"app"`
			} `json:"setSecrets"`
		}
		if err := client.graphQL(ctx, secretsSetMutation, variables, &result); err != nil {
			return fmt.Errorf("setting secret %q in app %q: %w", key, app, err)
		}

		secrets := result.SetSecrets.App.Secrets
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(secrets)
		}
		fmt.Printf("Set secret %q in app %s (%d secrets total)\n", key, app, len(secrets))
		return nil
	}
}

func newSecretsUnsetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset (remove) a secret from an app",
		RunE:  makeRunSecretsUnset(factory),
	}
	cmd.Flags().String("app", "", "App name / app ID (required)")
	cmd.Flags().StringSlice("keys", nil, "Secret key names to remove (required, comma-separated)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("keys")
	return cmd
}

func makeRunSecretsUnset(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		keys, _ := cmd.Flags().GetStringSlice("keys")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would unset secrets %v in app %q", keys, app), map[string]any{
				"action": "unset",
				"app":    app,
				"keys":   keys,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		variables := map[string]any{
			"appId": app,
			"keys":  keys,
		}

		var result struct {
			UnsetSecrets struct {
				App struct {
					Secrets []SecretSummary `json:"secrets"`
				} `json:"app"`
			} `json:"unsetSecrets"`
		}
		if err := client.graphQL(ctx, secretsUnsetMutation, variables, &result); err != nil {
			return fmt.Errorf("unsetting secrets in app %q: %w", app, err)
		}

		secrets := result.UnsetSecrets.App.Secrets
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(secrets)
		}
		fmt.Printf("Unset secrets %v from app %s (%d secrets remaining)\n", keys, app, len(secrets))
		return nil
	}
}
