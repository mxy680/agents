package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAnalyticsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "analytics",
		Aliases: []string{"logs"},
		Short:   "Analytics and logs",
	}
	cmd.AddCommand(
		newAnalyticsLogsCmd(factory),
		newAnalyticsAPICountsCmd(factory),
		newAnalyticsAPIRequestsCmd(factory),
		newAnalyticsFunctionsCmd(factory),
	)
	return cmd
}

// printAnalyticsJSON prints analytics data as pretty JSON or raw text.
func printAnalyticsJSON(cmd *cobra.Command, data []byte) error {
	if cli.IsJSONOutput(cmd) {
		var raw json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
		return cli.PrintJSON(raw)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(pretty.String())
	return nil
}

// --- logs ---

func newAnalyticsLogsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get analytics logs for a project",
		RunE:  makeRunAnalyticsEndpoint(factory, "logs.all"),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

// --- api-counts ---

func newAnalyticsAPICountsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-counts",
		Short: "Get API counts usage analytics for a project",
		RunE:  makeRunAnalyticsEndpoint(factory, "usage.api-counts"),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

// --- api-requests ---

func newAnalyticsAPIRequestsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-requests",
		Short: "Get API requests count analytics for a project",
		RunE:  makeRunAnalyticsEndpoint(factory, "usage.api-requests-count"),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

// --- functions ---

func newAnalyticsFunctionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "functions",
		Short: "Get Edge Functions combined stats analytics for a project",
		RunE:  makeRunAnalyticsEndpoint(factory, "functions.combined-stats"),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

// makeRunAnalyticsEndpoint creates a RunE function for a specific analytics endpoint.
func makeRunAnalyticsEndpoint(factory ClientFactory, endpoint string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet,
			fmt.Sprintf("/projects/%s/analytics/endpoints/%s", ref, endpoint), nil)
		if err != nil {
			return fmt.Errorf("getting analytics %s: %w", endpoint, err)
		}

		return printAnalyticsJSON(cmd, data)
	}
}
