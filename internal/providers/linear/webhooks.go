package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newWebhooksListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE:  makeRunWebhooksList(factory),
	}
	return cmd
}

func makeRunWebhooksList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query {
  webhooks {
    nodes {
      id
      url
      enabled
      team { name }
      createdAt
    }
  }
}`

		var resp struct {
			Webhooks struct {
				Nodes []struct {
					ID      string `json:"id"`
					URL     string `json:"url"`
					Enabled bool   `json:"enabled"`
					Team    *struct {
						Name string `json:"name"`
					} `json:"team"`
					CreatedAt string `json:"createdAt"`
				} `json:"nodes"`
			} `json:"webhooks"`
		}

		if err := client.graphQL(ctx, q, nil, &resp); err != nil {
			return fmt.Errorf("listing webhooks: %w", err)
		}

		webhooks := make([]WebhookSummary, 0, len(resp.Webhooks.Nodes))
		for _, n := range resp.Webhooks.Nodes {
			w := WebhookSummary{
				ID:        n.ID,
				URL:       n.URL,
				Enabled:   n.Enabled,
				CreatedAt: n.CreatedAt,
			}
			if n.Team != nil {
				w.Team = n.Team.Name
			}
			webhooks = append(webhooks, w)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(webhooks)
		}
		if len(webhooks) == 0 {
			fmt.Println("No webhooks found.")
			return nil
		}
		lines := make([]string, 0, len(webhooks)+1)
		lines = append(lines, fmt.Sprintf("%-28s  %-45s  %-8s  %s", "ID", "URL", "ENABLED", "TEAM"))
		for _, w := range webhooks {
			lines = append(lines, fmt.Sprintf("%-28s  %-45s  %-8v  %s",
				truncate(w.ID, 28), truncate(w.URL, 45), w.Enabled, truncate(w.Team, 25)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newWebhooksCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		RunE:  makeRunWebhooksCreate(factory),
	}
	cmd.Flags().String("url", "", "Webhook URL (required)")
	cmd.Flags().String("team", "", "Team ID (required)")
	cmd.Flags().StringSlice("resource-types", []string{"Issue"}, "Resource types to subscribe to")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without making changes")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func makeRunWebhooksCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		url, _ := cmd.Flags().GetString("url")
		teamID, _ := cmd.Flags().GetString("team")
		resourceTypes, _ := cmd.Flags().GetStringSlice("resource-types")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create webhook for %s", url), map[string]any{
				"action":        "create",
				"url":           url,
				"teamId":        teamID,
				"resourceTypes": resourceTypes,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($input: WebhookCreateInput!) {
  webhookCreate(input: $input) {
    webhook {
      id
      url
    }
  }
}`

		var resp struct {
			WebhookCreate struct {
				Webhook struct {
					ID  string `json:"id"`
					URL string `json:"url"`
				} `json:"webhook"`
			} `json:"webhookCreate"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"input": map[string]any{
			"url":           url,
			"teamId":        teamID,
			"resourceTypes": resourceTypes,
		}}, &resp); err != nil {
			return fmt.Errorf("creating webhook: %w", err)
		}

		webhook := resp.WebhookCreate.Webhook
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(webhook)
		}
		fmt.Printf("Created webhook: %s (ID: %s)\n", webhook.URL, webhook.ID)
		return nil
	}
}

func newWebhooksDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a webhook (irreversible)",
		RunE:  makeRunWebhooksDelete(factory),
	}
	cmd.Flags().String("id", "", "Webhook ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without making changes")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunWebhooksDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete webhook %q", id), map[string]any{
				"action": "delete",
				"id":     id,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($id: String!) {
  webhookDelete(id: $id) {
    success
  }
}`

		var resp struct {
			WebhookDelete struct {
				Success bool `json:"success"`
			} `json:"webhookDelete"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("deleting webhook %q: %w", id, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"success": resp.WebhookDelete.Success, "id": id})
		}
		fmt.Printf("Deleted webhook: %s\n", id)
		return nil
	}
}
