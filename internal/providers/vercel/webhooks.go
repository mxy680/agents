package vercel

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// WebhookSummary is the JSON-serializable representation of a Vercel webhook.
type WebhookSummary struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events,omitempty"`
	CreatedAt int64    `json:"createdAt,omitempty"`
}

func toWebhookSummary(data map[string]any) WebhookSummary {
	return WebhookSummary{
		ID:        jsonString(data["id"]),
		URL:       jsonString(data["url"]),
		Events:    jsonStringSlice(data["events"]),
		CreatedAt: jsonInt64(data["createdAt"]),
	}
}

func printWebhookSummaries(cmd *cobra.Command, hooks []WebhookSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(hooks)
	}
	if len(hooks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No webhooks found.")
		return nil
	}
	lines := make([]string, 0, len(hooks)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-50s  %s", "ID", "URL", "EVENTS"))
	for _, h := range hooks {
		lines = append(lines, fmt.Sprintf("%-28s  %-50s  %s",
			truncate(h.ID, 28), truncate(h.URL, 50), strings.Join(h.Events, ",")))
	}
	cli.PrintText(lines)
	return nil
}

func newWebhooksListCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE:  makeRunWebhooksList(factory),
	}
}

func makeRunWebhooksList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var hooks []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, "/v1/webhooks", nil, &hooks); err != nil {
			return fmt.Errorf("listing webhooks: %w", err)
		}

		summaries := make([]WebhookSummary, 0, len(hooks))
		for _, h := range hooks {
			summaries = append(summaries, toWebhookSummary(h))
		}

		return printWebhookSummaries(cmd, summaries)
	}
}

func newWebhooksCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		RunE:  makeRunWebhooksCreate(factory),
	}
	cmd.Flags().String("url", "", "Webhook endpoint URL (required)")
	cmd.Flags().String("events", "", "Comma-separated list of event types to subscribe to (required)")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("events")
	return cmd
}

func makeRunWebhooksCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		url, _ := cmd.Flags().GetString("url")
		eventsRaw, _ := cmd.Flags().GetString("events")

		events := strings.Split(eventsRaw, ",")
		for i, e := range events {
			events[i] = strings.TrimSpace(e)
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create webhook targeting %q", url), map[string]any{
				"action": "create",
				"url":    url,
				"events": events,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"url":    url,
			"events": events,
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, "/v1/webhooks", body, &data); err != nil {
			return fmt.Errorf("creating webhook: %w", err)
		}

		h := toWebhookSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(h)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created webhook: %s (ID: %s)\n", h.URL, h.ID)
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
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunWebhooksDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		webhookID, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete webhook %q", webhookID), map[string]any{
				"action":    "delete",
				"webhookId": webhookID,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/webhooks/%s", webhookID), nil); err != nil {
			return fmt.Errorf("deleting webhook %q: %w", webhookID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "webhookId": webhookID})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted webhook: %s\n", webhookID)
		return nil
	}
}
