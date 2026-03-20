package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newDomainsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domains",
		Aliases: []string{"domain"},
		Short:   "Custom hostname and vanity subdomain",
	}

	customCmd := &cobra.Command{
		Use:   "custom",
		Short: "Custom hostname management",
	}
	customCmd.AddCommand(
		newDomainsCustomGetCmd(factory),
		newDomainsCustomDeleteCmd(factory),
		newDomainsCustomInitializeCmd(factory),
		newDomainsCustomVerifyCmd(factory),
		newDomainsCustomActivateCmd(factory),
	)

	vanityCmd := &cobra.Command{
		Use:   "vanity",
		Short: "Vanity subdomain management",
	}
	vanityCmd.AddCommand(
		newDomainsVanityGetCmd(factory),
		newDomainsVanityDeleteCmd(factory),
		newDomainsVanityCheckCmd(factory),
		newDomainsVanityActivateCmd(factory),
	)

	cmd.AddCommand(customCmd, vanityCmd)
	return cmd
}

// --- custom get ---

func newDomainsCustomGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get custom hostname configuration",
		RunE:  makeRunDomainsCustomGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDomainsCustomGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/custom-hostname", ref), nil)
		if err != nil {
			return fmt.Errorf("getting custom hostname for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing custom hostname response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		hostname, _ := data["custom_hostname"].(string)
		status, _ := data["status"].(string)
		lines := []string{
			fmt.Sprintf("Hostname: %s", hostname),
			fmt.Sprintf("Status:   %s", status),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- custom delete ---

func newDomainsCustomDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete custom hostname configuration",
		RunE:  makeRunDomainsCustomDelete(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDomainsCustomDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would delete custom hostname for project %q", ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting custom hostname for project %q", ref)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/custom-hostname", ref), nil); err != nil {
			return fmt.Errorf("deleting custom hostname for project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "ref": ref})
		}
		fmt.Printf("Deleted custom hostname for project %q\n", ref)
		return nil
	}
}

// --- custom initialize ---

func newDomainsCustomInitializeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initialize",
		Short: "Initialize custom hostname configuration",
		RunE:  makeRunDomainsCustomInitialize(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("hostname", "", "Custom hostname (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("hostname")
	return cmd
}

func makeRunDomainsCustomInitialize(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		hostname, _ := cmd.Flags().GetString("hostname")

		if dryRunResult(cmd, fmt.Sprintf("Would initialize custom hostname %q for project %q", hostname, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{"custom_hostname": hostname}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		raw, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/custom-hostname/initialize", ref), strings.NewReader(string(bodyBytes)))
		if err != nil {
			return fmt.Errorf("initializing custom hostname for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing custom hostname response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		fmt.Printf("Initialized custom hostname %q for project %q\n", hostname, ref)
		return nil
	}
}

// --- custom verify ---

func newDomainsCustomVerifyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify (reverify) custom hostname",
		RunE:  makeRunDomainsCustomVerify(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDomainsCustomVerify(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would reverify custom hostname for project %q", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/custom-hostname/reverify", ref), nil)
		if err != nil {
			return fmt.Errorf("reverifying custom hostname for project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			var result any
			json.Unmarshal(raw, &result)
			return cli.PrintJSON(result)
		}

		fmt.Printf("Triggered reverification of custom hostname for project %q\n", ref)
		return nil
	}
}

// --- custom activate ---

func newDomainsCustomActivateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate custom hostname",
		RunE:  makeRunDomainsCustomActivate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDomainsCustomActivate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would activate custom hostname for project %q", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/custom-hostname/activate", ref), nil)
		if err != nil {
			return fmt.Errorf("activating custom hostname for project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			var result any
			json.Unmarshal(raw, &result)
			return cli.PrintJSON(result)
		}

		fmt.Printf("Activated custom hostname for project %q\n", ref)
		return nil
	}
}

// --- vanity get ---

func newDomainsVanityGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get vanity subdomain configuration",
		RunE:  makeRunDomainsVanityGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDomainsVanityGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/vanity-subdomain", ref), nil)
		if err != nil {
			return fmt.Errorf("getting vanity subdomain for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing vanity subdomain response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		subdomain, _ := data["vanity_subdomain"].(string)
		lines := []string{
			fmt.Sprintf("Vanity Subdomain: %s", subdomain),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- vanity delete ---

func newDomainsVanityDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete vanity subdomain",
		RunE:  makeRunDomainsVanityDelete(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDomainsVanityDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would delete vanity subdomain for project %q", ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting vanity subdomain for project %q", ref)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/vanity-subdomain", ref), nil); err != nil {
			return fmt.Errorf("deleting vanity subdomain for project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "ref": ref})
		}
		fmt.Printf("Deleted vanity subdomain for project %q\n", ref)
		return nil
	}
}

// --- vanity check ---

func newDomainsVanityCheckCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check availability of a vanity subdomain",
		RunE:  makeRunDomainsVanityCheck(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("subdomain", "", "Subdomain to check (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("subdomain")
	return cmd
}

func makeRunDomainsVanityCheck(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		subdomain, _ := cmd.Flags().GetString("subdomain")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{"vanity_subdomain": subdomain}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		raw, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/vanity-subdomain/check-availability", ref), strings.NewReader(string(bodyBytes)))
		if err != nil {
			return fmt.Errorf("checking vanity subdomain availability for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing availability response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		available, _ := data["available"].(bool)
		lines := []string{
			fmt.Sprintf("Subdomain:  %s", subdomain),
			fmt.Sprintf("Available:  %v", available),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- vanity activate ---

func newDomainsVanityActivateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate a vanity subdomain",
		RunE:  makeRunDomainsVanityActivate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("subdomain", "", "Subdomain to activate (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("subdomain")
	return cmd
}

func makeRunDomainsVanityActivate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		subdomain, _ := cmd.Flags().GetString("subdomain")

		if dryRunResult(cmd, fmt.Sprintf("Would activate vanity subdomain %q for project %q", subdomain, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{"vanity_subdomain": subdomain}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		raw, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/vanity-subdomain/activate", ref), strings.NewReader(string(bodyBytes)))
		if err != nil {
			return fmt.Errorf("activating vanity subdomain for project %q: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			var result any
			json.Unmarshal(raw, &result)
			return cli.PrintJSON(result)
		}

		fmt.Printf("Activated vanity subdomain %q for project %q\n", subdomain, ref)
		return nil
	}
}
