package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAdvisorsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "advisors",
		Aliases: []string{"advisor"},
		Short:   "Performance and security advisors",
	}
	cmd.AddCommand(newAdvisorsPerformanceCmd(factory), newAdvisorsSecurityCmd(factory))
	return cmd
}

// printAdvisorResults outputs advisor results as JSON or a formatted text table.
func printAdvisorResults(cmd *cobra.Command, results []AdvisorResult) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(results)
	}
	if len(results) == 0 {
		fmt.Println("No advisor results found.")
		return nil
	}
	lines := make([]string, 0, len(results)+1)
	lines = append(lines, fmt.Sprintf("%-10s  %-40s  %s", "SEVERITY", "TITLE", "DESCRIPTION"))
	for _, r := range results {
		lines = append(lines, fmt.Sprintf("%-10s  %-40s  %s",
			truncate(r.Severity, 10), truncate(r.Title, 40), truncate(r.Description, 80)))
	}
	cli.PrintText(lines)
	return nil
}

// --- performance ---

func newAdvisorsPerformanceCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "performance",
		Short: "Get performance advisor results for a project",
		RunE:  makeRunAdvisors(factory, "performance"),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

// --- security ---

func newAdvisorsSecurityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Get security advisor results for a project",
		RunE:  makeRunAdvisors(factory, "security"),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

// makeRunAdvisors creates a RunE function for a specific advisor type.
func makeRunAdvisors(factory ClientFactory, advisorType string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet,
			fmt.Sprintf("/projects/%s/advisors/%s", ref, advisorType), nil)
		if err != nil {
			return fmt.Errorf("getting %s advisor results: %w", advisorType, err)
		}

		var results []AdvisorResult
		if err := json.Unmarshal(data, &results); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		return printAdvisorResults(cmd, results)
	}
}
