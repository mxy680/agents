package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newZonesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List zones",
		RunE:  makeRunZonesList(factory),
	}
	cmd.Flags().Int("per-page", 50, "Number of zones per page")
	return cmd
}

func makeRunZonesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		perPage, _ := cmd.Flags().GetInt("per-page")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/zones?per_page=%d", perPage), nil, &resp); err != nil {
			return fmt.Errorf("listing zones: %w", err)
		}

		summaries := make([]ZoneSummary, 0, len(resp))
		for _, z := range resp {
			summaries = append(summaries, toZoneSummary(z))
		}

		return printZoneSummaries(cmd, summaries)
	}
}

func newZonesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get zone details",
		RunE:  makeRunZonesGet(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	_ = cmd.MarkFlagRequired("zone")
	return cmd
}

func makeRunZonesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/zones/%s", zoneID), nil, &data); err != nil {
			return fmt.Errorf("getting zone %q: %w", zoneID, err)
		}

		detail := toZoneDetail(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:           %s", detail.ID),
			fmt.Sprintf("Name:         %s", detail.Name),
			fmt.Sprintf("Status:       %s", detail.Status),
			fmt.Sprintf("Plan:         %s", detail.Plan),
			fmt.Sprintf("Type:         %s", detail.Type),
			fmt.Sprintf("Paused:       %v", detail.Paused),
		}
		for _, ns := range detail.NameServers {
			lines = append(lines, fmt.Sprintf("Name Server:  %s", ns))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newZonesPurgeCacheCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge-cache",
		Short: "Purge zone cache",
		RunE:  makeRunZonesPurgeCache(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().StringSlice("files", nil, "Specific file URLs to purge (omit to purge everything)")
	cmd.Flags().Bool("confirm", false, "Confirm cache purge")
	_ = cmd.MarkFlagRequired("zone")
	return cmd
}

func makeRunZonesPurgeCache(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		files, _ := cmd.Flags().GetStringSlice("files")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would purge cache for zone %q", zoneID), map[string]any{
				"action":  "purge-cache",
				"zone_id": zoneID,
				"files":   files,
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

		var body map[string]any
		if len(files) > 0 {
			body = map[string]any{"files": files}
		} else {
			body = map[string]any{"purge_everything": true}
		}

		if _, err := client.do(ctx, http.MethodPost, fmt.Sprintf("/zones/%s/purge_cache", zoneID), body); err != nil {
			return fmt.Errorf("purging cache for zone %q: %w", zoneID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "purged", "zone_id": zoneID})
		}
		fmt.Printf("Cache purged for zone: %s\n", zoneID)
		return nil
	}
}
