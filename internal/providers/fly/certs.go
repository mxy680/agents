package fly

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newCertsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certificates for an app",
		RunE:  makeRunCertsList(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	_ = cmd.MarkFlagRequired("app")
	return cmd
}

func makeRunCertsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var certs []CertSummary
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s/certificates", app), nil, &certs); err != nil {
			return fmt.Errorf("listing certificates for app %q: %w", app, err)
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
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("hostname", "", "Hostname (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("hostname")
	return cmd
}

func makeRunCertsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		hostname, _ := cmd.Flags().GetString("hostname")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data CertDetail
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s/certificates/%s", app, hostname), nil, &data); err != nil {
			return fmt.Errorf("getting certificate for %q: %w", hostname, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		lines := []string{
			fmt.Sprintf("Hostname:              %s", data.Hostname),
			fmt.Sprintf("Client Status:         %s", data.ClientStatus),
			fmt.Sprintf("Issued:                %v", data.Issued),
			fmt.Sprintf("ACME DNS Configured:   %v", data.AcmeDNSConfigured),
			fmt.Sprintf("ACME ALPN Configured:  %v", data.AcmeALPNConfigured),
			fmt.Sprintf("DNS Validation Target: %s", data.DNSValidationTarget),
			fmt.Sprintf("Created At:            %s", data.CreatedAt),
			fmt.Sprintf("Expires At:            %s", data.ExpiresAt),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newCertsAddCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a certificate for a hostname",
		RunE:  makeRunCertsAdd(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("hostname", "", "Hostname to add certificate for (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("hostname")
	return cmd
}

func makeRunCertsAdd(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		hostname, _ := cmd.Flags().GetString("hostname")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would add certificate for %q in app %q", hostname, app), map[string]any{
				"action":   "add",
				"app":      app,
				"hostname": hostname,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"hostname": hostname}
		var data CertDetail
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/apps/%s/certificates/acme", app), body, &data); err != nil {
			return fmt.Errorf("adding certificate for %q: %w", hostname, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Added certificate for hostname: %s (status: %s)\n", data.Hostname, data.ClientStatus)
		return nil
	}
}

func newCertsCheckCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check certificate status for a hostname",
		RunE:  makeRunCertsCheck(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("hostname", "", "Hostname (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("hostname")
	return cmd
}

func makeRunCertsCheck(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		hostname, _ := cmd.Flags().GetString("hostname")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data CertDetail
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/apps/%s/certificates/%s/check", app, hostname), nil, &data); err != nil {
			return fmt.Errorf("checking certificate for %q: %w", hostname, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Certificate check for %s: status=%s, issued=%v\n", data.Hostname, data.ClientStatus, data.Issued)
		return nil
	}
}

func newCertsRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a certificate for a hostname (irreversible)",
		RunE:  makeRunCertsRemove(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("hostname", "", "Hostname (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible removal")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("hostname")
	return cmd
}

func makeRunCertsRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		hostname, _ := cmd.Flags().GetString("hostname")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would remove certificate for %q in app %q", hostname, app), map[string]any{
				"action":   "remove",
				"app":      app,
				"hostname": hostname,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/apps/%s/certificates/%s", app, hostname), nil); err != nil {
			return fmt.Errorf("removing certificate for %q: %w", hostname, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "removed", "hostname": hostname})
		}
		fmt.Printf("Removed certificate for: %s\n", hostname)
		return nil
	}
}
