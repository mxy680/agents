package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newDNSListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records for a zone",
		RunE:  makeRunDNSList(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("type", "", "Filter by record type (e.g. A, CNAME, MX)")
	cmd.Flags().Int("per-page", 100, "Number of records per page")
	_ = cmd.MarkFlagRequired("zone")
	return cmd
}

func makeRunDNSList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		recordType, _ := cmd.Flags().GetString("type")
		perPage, _ := cmd.Flags().GetInt("per-page")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/zones/%s/dns_records?per_page=%d", zoneID, perPage)
		if recordType != "" {
			path += "&type=" + recordType
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing DNS records for zone %q: %w", zoneID, err)
		}

		records := make([]DNSRecordSummary, 0, len(resp))
		for _, r := range resp {
			records = append(records, toDNSRecordSummary(r))
		}

		return printDNSRecords(cmd, records)
	}
}

func newDNSGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a DNS record",
		RunE:  makeRunDNSGet(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("record", "", "DNS record ID (required)")
	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("record")
	return cmd
}

func makeRunDNSGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		recordID, _ := cmd.Flags().GetString("record")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID), nil, &data); err != nil {
			return fmt.Errorf("getting DNS record %q: %w", recordID, err)
		}

		record := toDNSRecordSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(record)
		}

		lines := []string{
			fmt.Sprintf("ID:      %s", record.ID),
			fmt.Sprintf("Type:    %s", record.Type),
			fmt.Sprintf("Name:    %s", record.Name),
			fmt.Sprintf("Content: %s", record.Content),
			fmt.Sprintf("Proxied: %v", record.Proxied),
			fmt.Sprintf("TTL:     %d", record.TTL),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newDNSCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a DNS record",
		RunE:  makeRunDNSCreate(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("type", "", "Record type (e.g. A, CNAME, MX) (required)")
	cmd.Flags().String("name", "", "Record name (required)")
	cmd.Flags().String("content", "", "Record content/value (required)")
	cmd.Flags().Int("ttl", 1, "TTL in seconds (1 = auto)")
	cmd.Flags().Bool("proxied", false, "Proxy through Cloudflare")
	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

func makeRunDNSCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		recordType, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")
		content, _ := cmd.Flags().GetString("content")
		ttl, _ := cmd.Flags().GetInt("ttl")
		proxied, _ := cmd.Flags().GetBool("proxied")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create %s record %q → %q", recordType, name, content), map[string]any{
				"action":  "create",
				"zone_id": zoneID,
				"type":    recordType,
				"name":    name,
				"content": content,
				"ttl":     ttl,
				"proxied": proxied,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"type":    recordType,
			"name":    name,
			"content": content,
			"ttl":     ttl,
			"proxied": proxied,
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/zones/%s/dns_records", zoneID), body, &data); err != nil {
			return fmt.Errorf("creating DNS record: %w", err)
		}

		record := toDNSRecordSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(record)
		}
		fmt.Printf("Created DNS record: %s %s → %s (ID: %s)\n", record.Type, record.Name, record.Content, record.ID)
		return nil
	}
}

func newDNSUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a DNS record",
		RunE:  makeRunDNSUpdate(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("record", "", "DNS record ID (required)")
	cmd.Flags().String("type", "", "Record type")
	cmd.Flags().String("name", "", "Record name")
	cmd.Flags().String("content", "", "Record content/value")
	cmd.Flags().Int("ttl", 0, "TTL in seconds (0 = leave unchanged)")
	cmd.Flags().Bool("proxied", false, "Proxy through Cloudflare")
	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("record")
	return cmd
}

func makeRunDNSUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		recordID, _ := cmd.Flags().GetString("record")
		recordType, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")
		content, _ := cmd.Flags().GetString("content")
		ttl, _ := cmd.Flags().GetInt("ttl")
		proxied, _ := cmd.Flags().GetBool("proxied")

		body := map[string]any{}
		if recordType != "" {
			body["type"] = recordType
		}
		if name != "" {
			body["name"] = name
		}
		if content != "" {
			body["content"] = content
		}
		if ttl != 0 {
			body["ttl"] = ttl
		}
		if cmd.Flags().Changed("proxied") {
			body["proxied"] = proxied
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update DNS record %q", recordID), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID), body, &data); err != nil {
			return fmt.Errorf("updating DNS record %q: %w", recordID, err)
		}

		record := toDNSRecordSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(record)
		}
		fmt.Printf("Updated DNS record: %s %s → %s\n", record.Type, record.Name, record.Content)
		return nil
	}
}

func newDNSDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a DNS record (irreversible)",
		RunE:  makeRunDNSDelete(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("record", "", "DNS record ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("record")
	return cmd
}

func makeRunDNSDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		recordID, _ := cmd.Flags().GetString("record")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete DNS record %q from zone %q", recordID, zoneID), map[string]any{
				"action":    "delete",
				"zone_id":   zoneID,
				"record_id": recordID,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID), nil); err != nil {
			return fmt.Errorf("deleting DNS record %q: %w", recordID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "record_id": recordID})
		}
		fmt.Printf("Deleted DNS record: %s\n", recordID)
		return nil
	}
}
