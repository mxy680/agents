package imessage

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newWebhooksCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhooks",
		Short:   "Manage webhooks",
		Aliases: []string{"webhook", "wh"},
	}

	cmd.AddCommand(newWebhooksListCmd(factory))
	cmd.AddCommand(newWebhooksCreateCmd(factory))
	cmd.AddCommand(newWebhooksDeleteCmd(factory))

	return cmd
}

func newWebhooksListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunWebhooksList(factory)
	return cmd
}

func makeRunWebhooksList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "webhook", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var webhooks []WebhookSummary
		if err := json.Unmarshal(data, &webhooks); err != nil {
			return printResult(cmd, data, []string{string(data)})
		}

		lines := make([]string, 0, len(webhooks))
		for _, w := range webhooks {
			events := ""
			if len(w.Events) > 0 {
				events = fmt.Sprintf("  [%s]", joinStrings(w.Events, ", "))
			}
			lines = append(lines, fmt.Sprintf("%-4d  %-50s%s", w.ID, truncate(w.URL, 48), events))
		}
		if len(lines) == 0 {
			lines = []string{"No webhooks configured."}
		}

		return printResult(cmd, webhooks, lines)
	}
}

func newWebhooksCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
	}
	cmd.Flags().String("url", "", "Webhook URL (required)")
	cmd.Flags().String("events", "", "Comma-separated list of events to subscribe to (optional)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("url")
	cmd.RunE = makeRunWebhooksCreate(factory)
	return cmd
}

func makeRunWebhooksCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		webhookURL, _ := cmd.Flags().GetString("url")
		eventsRaw, _ := cmd.Flags().GetString("events")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		var events []string
		if eventsRaw != "" {
			events = splitCSV(eventsRaw)
		}

		if dryRun {
			result := dryRunResult("webhooks create", map[string]any{
				"url":    webhookURL,
				"events": events,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would create webhook for: %s", webhookURL)})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		body := map[string]any{
			"url":    webhookURL,
			"events": events,
		}

		resp, err := client.Post(cmd.Context(), "webhook", body)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var w WebhookSummary
		if err := json.Unmarshal(data, &w); err != nil {
			return printResult(cmd, data, []string{fmt.Sprintf("Webhook created for: %s", webhookURL)})
		}

		return printResult(cmd, w, []string{fmt.Sprintf("Webhook created (ID: %d): %s", w.ID, w.URL)})
	}
}

func newWebhooksDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a webhook",
	}
	cmd.Flags().String("id", "", "Webhook ID to delete (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("id")
	cmd.RunE = makeRunWebhooksDelete(factory)
	return cmd
}

func makeRunWebhooksDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("webhooks delete", map[string]any{"id": id})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would delete webhook: %s", id)})
		}

		if err := confirmDestructive(cmd, "delete webhook"); err != nil {
			return err
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Delete(cmd.Context(), fmt.Sprintf("webhook/%s", id))
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{fmt.Sprintf("Webhook %s deleted.", id)})
	}
}

// joinStrings joins a slice with a separator.
func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
