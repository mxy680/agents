package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newCertsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSL certificate packs for a zone",
		RunE:  makeRunCertsList(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	_ = cmd.MarkFlagRequired("zone")
	return cmd
}

func makeRunCertsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/zones/%s/ssl/certificate_packs", zoneID), nil, &resp); err != nil {
			return fmt.Errorf("listing certificate packs for zone %q: %w", zoneID, err)
		}

		packs := make([]CertPackSummary, 0, len(resp))
		for _, p := range resp {
			packs = append(packs, toCertPackSummary(p))
		}

		return printCertPacks(cmd, packs)
	}
}

func newCertsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an SSL certificate pack",
		RunE:  makeRunCertsGet(factory),
	}
	cmd.Flags().String("zone", "", "Zone ID (required)")
	cmd.Flags().String("cert", "", "Certificate pack ID (required)")
	_ = cmd.MarkFlagRequired("zone")
	_ = cmd.MarkFlagRequired("cert")
	return cmd
}

func makeRunCertsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		zoneID, _ := cmd.Flags().GetString("zone")
		certID, _ := cmd.Flags().GetString("cert")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/zones/%s/ssl/certificate_packs/%s", zoneID, certID), nil, &data); err != nil {
			return fmt.Errorf("getting certificate pack %q: %w", certID, err)
		}

		pack := toCertPackSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(pack)
		}

		lines := []string{
			fmt.Sprintf("ID:     %s", pack.ID),
			fmt.Sprintf("Type:   %s", pack.Type),
			fmt.Sprintf("Status: %s", pack.Status),
		}
		for _, h := range pack.Hosts {
			lines = append(lines, fmt.Sprintf("Host:   %s", h))
		}
		cli.PrintText(lines)
		return nil
	}
}
