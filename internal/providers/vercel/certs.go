package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// CertSummary is the JSON-serializable representation of a Vercel certificate.
type CertSummary struct {
	ID         string   `json:"id"`
	CNs        []string `json:"cns,omitempty"`
	Expiration int64    `json:"expiration,omitempty"`
	CreatedAt  int64    `json:"createdAt,omitempty"`
	AutoRenew  bool     `json:"autoRenew,omitempty"`
}

func toCertSummary(data map[string]any) CertSummary {
	return CertSummary{
		ID:         jsonString(data["id"]),
		CNs:        jsonStringSlice(data["cns"]),
		Expiration: jsonInt64(data["expiration"]),
		CreatedAt:  jsonInt64(data["createdAt"]),
		AutoRenew:  jsonBool(data["autoRenew"]),
	}
}

func printCertSummaries(cmd *cobra.Command, certs []CertSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(certs)
	}
	if len(certs) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No certificates found.")
		return nil
	}
	lines := make([]string, 0, len(certs)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-10s  %s", "ID", "AUTO-RENEW", "CNS"))
	for _, c := range certs {
		cns := ""
		if len(c.CNs) > 0 {
			cns = c.CNs[0]
			if len(c.CNs) > 1 {
				cns += fmt.Sprintf(" (+%d more)", len(c.CNs)-1)
			}
		}
		lines = append(lines, fmt.Sprintf("%-28s  %-10v  %s", truncate(c.ID, 28), c.AutoRenew, cns))
	}
	cli.PrintText(lines)
	return nil
}

func newCertsListCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List TLS certificates",
		RunE:  makeRunCertsList(factory),
	}
}

func makeRunCertsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp struct {
			Certs []map[string]any `json:"certs"`
		}
		if err := client.doJSON(ctx, http.MethodGet, "/v4/certs", nil, &resp); err != nil {
			return fmt.Errorf("listing certificates: %w", err)
		}

		certs := make([]CertSummary, 0, len(resp.Certs))
		for _, c := range resp.Certs {
			certs = append(certs, toCertSummary(c))
		}

		return printCertSummaries(cmd, certs)
	}
}

func newCertsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get certificate details",
		RunE:  makeRunCertsGet(factory),
	}
	cmd.Flags().String("id", "", "Certificate ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCertsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		certID, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v4/certs/%s", certID), nil, &data); err != nil {
			return fmt.Errorf("getting certificate %q: %w", certID, err)
		}

		c := toCertSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(c)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", c.ID),
			fmt.Sprintf("Auto-Renew:  %v", c.AutoRenew),
		}
		for i, cn := range c.CNs {
			if i == 0 {
				lines = append(lines, fmt.Sprintf("CNs:         %s", cn))
			} else {
				lines = append(lines, fmt.Sprintf("             %s", cn))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}
