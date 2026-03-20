package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newOrgsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "orgs",
		Aliases: []string{"org"},
		Short:   "Manage Supabase organizations",
	}
	cmd.AddCommand(
		newOrgsListCmd(factory),
		newOrgsCreateCmd(factory),
	)
	return cmd
}

// --- Converters ---

func toOrgSummary(data map[string]any) OrgSummary {
	s := func(key string) string {
		v, _ := data[key].(string)
		return v
	}
	return OrgSummary{
		ID:   s("id"),
		Name: s("name"),
	}
}

// --- Commands ---

func newOrgsListCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all organizations",
		RunE:  makeRunOrgsList(factory),
	}
}

func makeRunOrgsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, "/organizations", nil)
		if err != nil {
			return fmt.Errorf("listing organizations: %w", err)
		}

		var data []map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing organizations response: %w", err)
		}

		summaries := make([]OrgSummary, 0, len(data))
		for _, d := range data {
			summaries = append(summaries, toOrgSummary(d))
		}
		return printOrgSummaries(cmd, summaries)
	}
}

func newOrgsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new organization",
		RunE:  makeRunOrgsCreate(factory),
	}
	cmd.Flags().String("name", "", "Organization name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunOrgsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if dryRunResult(cmd, fmt.Sprintf("Would create organization %q", name)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"name": name}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		raw, err := doSupabase(client, http.MethodPost, "/organizations", bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("creating organization %q: %w", name, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing create response: %w", err)
		}

		org := toOrgSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(org)
		}
		fmt.Printf("Created: %s (%s)\n", org.Name, org.ID)
		return nil
	}
}
