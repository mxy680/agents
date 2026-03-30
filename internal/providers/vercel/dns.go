package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// DNSRecord is the JSON-serializable representation of a Vercel DNS record.
type DNSRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	TTL       int64  `json:"ttl,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
}

func toDNSRecord(data map[string]any) DNSRecord {
	return DNSRecord{
		ID:        jsonString(data["id"]),
		Name:      jsonString(data["name"]),
		Type:      jsonString(data["type"]),
		Value:     jsonString(data["value"]),
		TTL:       jsonInt64(data["ttl"]),
		CreatedAt: jsonInt64(data["createdAt"]),
		UpdatedAt: jsonInt64(data["updatedAt"]),
	}
}

func printDNSRecords(cmd *cobra.Command, records []DNSRecord) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(records)
	}
	if len(records) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No DNS records found.")
		return nil
	}
	lines := make([]string, 0, len(records)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-40s  %-8s  %s", "ID", "NAME", "TYPE", "VALUE"))
	for _, r := range records {
		lines = append(lines, fmt.Sprintf("%-28s  %-40s  %-8s  %s",
			truncate(r.ID, 28), truncate(r.Name, 40), r.Type, truncate(r.Value, 50)))
	}
	cli.PrintText(lines)
	return nil
}

func newDNSListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records for a domain",
		RunE:  makeRunDNSList(factory),
	}
	cmd.Flags().String("domain", "", "Domain name (required)")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func makeRunDNSList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		domain, _ := cmd.Flags().GetString("domain")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp struct {
			Records []map[string]any `json:"records"`
		}
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v4/domains/%s/records", domain), nil, &resp); err != nil {
			return fmt.Errorf("listing DNS records for domain %q: %w", domain, err)
		}

		records := make([]DNSRecord, 0, len(resp.Records))
		for _, r := range resp.Records {
			records = append(records, toDNSRecord(r))
		}

		return printDNSRecords(cmd, records)
	}
}

func newDNSAddCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a DNS record to a domain",
		RunE:  makeRunDNSAdd(factory),
	}
	cmd.Flags().String("domain", "", "Domain name (required)")
	cmd.Flags().String("type", "", "Record type: A, AAAA, CNAME, MX, TXT (required)")
	cmd.Flags().String("name", "", "Record name / subdomain (required)")
	cmd.Flags().String("value", "", "Record value (required)")
	cmd.Flags().Int64("ttl", 0, "Time to live in seconds (optional, uses Vercel default if 0)")
	_ = cmd.MarkFlagRequired("domain")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func makeRunDNSAdd(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		recType, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")
		value, _ := cmd.Flags().GetString("value")
		ttl, _ := cmd.Flags().GetInt64("ttl")

		body := map[string]any{
			"name":  name,
			"type":  recType,
			"value": value,
		}
		if ttl > 0 {
			body["ttl"] = ttl
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would add %s record %q to domain %q", recType, name, domain), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v2/domains/%s/records", domain), body, &data); err != nil {
			return fmt.Errorf("adding DNS record to domain %q: %w", domain, err)
		}

		rec := toDNSRecord(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(rec)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added DNS record: %s %s → %s (ID: %s)\n", rec.Type, rec.Name, rec.Value, rec.ID)
		return nil
	}
}

func newDNSRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete a DNS record (irreversible)",
		RunE:  makeRunDNSRemove(factory),
	}
	cmd.Flags().String("domain", "", "Domain name (required)")
	cmd.Flags().String("id", "", "DNS record ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("domain")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunDNSRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		recordID, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete DNS record %q from domain %q", recordID, domain), map[string]any{
				"action":   "remove",
				"domain":   domain,
				"recordId": recordID,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v2/domains/%s/records/%s", domain, recordID), nil); err != nil {
			return fmt.Errorf("deleting DNS record %q from domain %q: %w", recordID, domain, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "recordId": recordID})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted DNS record: %s\n", recordID)
		return nil
	}
}
