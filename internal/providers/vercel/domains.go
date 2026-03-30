package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newDomainsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		RunE:  makeRunDomainsList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of domains to return")
	cmd.Flags().String("page-token", "", "Pagination cursor (until timestamp)")
	return cmd
}

func makeRunDomainsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/v5/domains?limit=%d", limit)
		if pageToken != "" {
			path += "&until=" + pageToken
		}

		var resp struct {
			Domains    []map[string]any `json:"domains"`
			Pagination struct {
				Next int64 `json:"next"`
			} `json:"pagination"`
		}
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing domains: %w", err)
		}

		summaries := make([]DomainSummary, 0, len(resp.Domains))
		for _, d := range resp.Domains {
			summaries = append(summaries, toDomainSummary(d))
		}

		if resp.Pagination.Next != 0 && !cli.IsJSONOutput(cmd) {
			warnf("more results available — use --page-token=%d to fetch next page", resp.Pagination.Next)
		}

		return printDomainSummaries(cmd, summaries)
	}
}

func newDomainsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get domain details",
		RunE:  makeRunDomainsGet(factory),
	}
	cmd.Flags().String("domain", "", "Domain name (required)")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func makeRunDomainsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		domain, _ := cmd.Flags().GetString("domain")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var resp struct {
			Domain map[string]any `json:"domain"`
		}
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v5/domains/%s", domain), nil, &resp); err != nil {
			return fmt.Errorf("getting domain %q: %w", domain, err)
		}

		detail := toDomainDetail(resp.Domain)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Name:         %s", detail.Name),
			fmt.Sprintf("ID:           %s", detail.ID),
			fmt.Sprintf("Verified:     %v", detail.Verified),
			fmt.Sprintf("Service Type: %s", detail.ServiceType),
		}
		if len(detail.Nameservers) > 0 {
			for i, ns := range detail.Nameservers {
				if i == 0 {
					lines = append(lines, fmt.Sprintf("Nameservers:  %s", ns))
				} else {
					lines = append(lines, fmt.Sprintf("              %s", ns))
				}
			}
		}
		if len(detail.IntendedNS) > 0 {
			for i, ns := range detail.IntendedNS {
				if i == 0 {
					lines = append(lines, fmt.Sprintf("Intended NS:  %s", ns))
				} else {
					lines = append(lines, fmt.Sprintf("              %s", ns))
				}
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

func newDomainsAddCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a domain",
		RunE:  makeRunDomainsAdd(factory),
	}
	cmd.Flags().String("domain", "", "Domain name to add (required)")
	cmd.Flags().String("project", "", "Project to link the domain to (optional)")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func makeRunDomainsAdd(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		project, _ := cmd.Flags().GetString("project")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would add domain %q", domain), map[string]any{
				"action":  "add",
				"domain":  domain,
				"project": project,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"name": domain}
		if project != "" {
			body["project"] = project
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, "/v5/domains", body, &data); err != nil {
			return fmt.Errorf("adding domain %q: %w", domain, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Added domain: %s\n", domain)
		if v, ok := data["verified"].(bool); ok {
			fmt.Printf("Verified: %v\n", v)
		}
		return nil
	}
}

func newDomainsVerifyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify domain ownership",
		RunE:  makeRunDomainsVerify(factory),
	}
	cmd.Flags().String("domain", "", "Domain name to verify (required)")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func makeRunDomainsVerify(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		domain, _ := cmd.Flags().GetString("domain")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would trigger verification for domain %q", domain), map[string]any{
				"action": "verify",
				"domain": domain,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v5/domains/%s/verify", domain), nil, &data); err != nil {
			return fmt.Errorf("verifying domain %q: %w", domain, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		verified, _ := data["verified"].(bool)
		if verified {
			fmt.Printf("Domain %q is verified.\n", domain)
		} else {
			fmt.Printf("Domain %q verification pending — DNS changes may take time to propagate.\n", domain)
		}
		return nil
	}
}

func newDomainsRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a domain (irreversible)",
		RunE:  makeRunDomainsRemove(factory),
	}
	cmd.Flags().String("domain", "", "Domain name to remove (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible removal")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func makeRunDomainsRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		domain, _ := cmd.Flags().GetString("domain")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would remove domain %q", domain), map[string]any{
				"action": "remove",
				"domain": domain,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v5/domains/%s", domain), nil); err != nil {
			return fmt.Errorf("removing domain %q: %w", domain, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "removed", "domain": domain})
		}
		fmt.Printf("Removed domain: %s\n", domain)
		return nil
	}
}
