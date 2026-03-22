package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newOutcomesCmd returns the parent "outcomes" command with all subcommands attached.
func newOutcomesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "outcomes",
		Short:   "Manage Canvas learning outcomes",
		Aliases: []string{"outcome"},
	}

	cmd.AddCommand(newOutcomesListCmd(factory))
	cmd.AddCommand(newOutcomesGetCmd(factory))
	cmd.AddCommand(newOutcomesCreateCmd(factory))
	cmd.AddCommand(newOutcomesUpdateCmd(factory))
	cmd.AddCommand(newOutcomesDeleteCmd(factory))
	cmd.AddCommand(newOutcomesGroupsCmd(factory))
	cmd.AddCommand(newOutcomesResultsCmd(factory))

	return cmd
}

func newOutcomesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List outcomes for an account or course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			contextType, _ := cmd.Flags().GetString("context-type")
			contextID, _ := cmd.Flags().GetString("context-id")

			var path string
			switch strings.ToLower(contextType) {
			case "account":
				if contextID == "" {
					return fmt.Errorf("--context-id is required when --context-type=Account")
				}
				path = "/accounts/" + contextID + "/outcome_groups/root/outcomes"
			case "course", "":
				if contextID == "" {
					return fmt.Errorf("--context-id is required")
				}
				path = "/courses/" + contextID + "/outcome_groups/root/outcomes"
			default:
				return fmt.Errorf("--context-type must be Account or Course")
			}

			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var outcomes []map[string]any
			if err := json.Unmarshal(data, &outcomes); err != nil {
				return fmt.Errorf("parse outcomes: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(outcomes)
			}

			if len(outcomes) == 0 {
				fmt.Println("No outcomes found.")
				return nil
			}
			for _, o := range outcomes {
				id, _ := o["outcome_id"].(float64)
				title := ""
				if inner, ok := o["outcome"].(map[string]any); ok {
					title, _ = inner["title"].(string)
				}
				fmt.Printf("%-6.0f  %s\n", id, truncate(title, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("context-type", "Course", "Context type: Account or Course")
	cmd.Flags().String("context-id", "", "ID of the account or course (required)")
	return cmd
}

func newOutcomesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific outcome",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			outcomeID, _ := cmd.Flags().GetString("outcome-id")
			if outcomeID == "" {
				return fmt.Errorf("--outcome-id is required")
			}

			data, err := client.Get(ctx, "/outcomes/"+outcomeID, nil)
			if err != nil {
				return err
			}

			var outcome OutcomeSummary
			if err := json.Unmarshal(data, &outcome); err != nil {
				return fmt.Errorf("parse outcome: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(outcome)
			}

			fmt.Printf("ID:             %d\n", outcome.ID)
			fmt.Printf("Title:          %s\n", outcome.Title)
			fmt.Printf("Mastery Points: %.1f\n", outcome.MasteryPoints)
			if outcome.ContextType != "" {
				fmt.Printf("Context:        %s %d\n", outcome.ContextType, outcome.ContextID)
			}
			if outcome.Description != "" {
				fmt.Printf("Description:    %s\n", truncate(outcome.Description, 200))
			}
			return nil
		},
	}

	cmd.Flags().String("outcome-id", "", "Canvas outcome ID (required)")
	return cmd
}

func newOutcomesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new outcome and link it to a context",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			title, _ := cmd.Flags().GetString("title")
			if title == "" {
				return fmt.Errorf("--title is required")
			}

			contextType, _ := cmd.Flags().GetString("context-type")
			contextID, _ := cmd.Flags().GetString("context-id")

			var linkPath string
			switch strings.ToLower(contextType) {
			case "account":
				if contextID == "" {
					return fmt.Errorf("--context-id is required when --context-type=Account")
				}
				linkPath = "/accounts/" + contextID + "/outcome_groups/root/outcomes"
			case "course", "":
				if contextID == "" {
					return fmt.Errorf("--context-id is required")
				}
				linkPath = "/courses/" + contextID + "/outcome_groups/root/outcomes"
			default:
				return fmt.Errorf("--context-type must be Account or Course")
			}

			body := map[string]any{"title": title}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				body["description"] = v
			}
			if cmd.Flags().Changed("mastery-points") {
				v, _ := cmd.Flags().GetFloat64("mastery-points")
				body["mastery_points"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create outcome %q in %s %s", title, contextType, contextID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, linkPath, body)
			if err != nil {
				return err
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse created outcome: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}

			id, _ := result["outcome_id"].(float64)
			fmt.Printf("Outcome created: %.0f — %s\n", id, title)
			return nil
		},
	}

	cmd.Flags().String("title", "", "Outcome title (required)")
	cmd.Flags().String("description", "", "Outcome description")
	cmd.Flags().Float64("mastery-points", 0, "Points required for mastery")
	cmd.Flags().String("context-type", "Course", "Context type: Account or Course")
	cmd.Flags().String("context-id", "", "ID of the account or course (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newOutcomesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing outcome",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			outcomeID, _ := cmd.Flags().GetString("outcome-id")
			if outcomeID == "" {
				return fmt.Errorf("--outcome-id is required")
			}

			body := map[string]any{}
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				body["title"] = v
			}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				body["description"] = v
			}
			if cmd.Flags().Changed("mastery-points") {
				v, _ := cmd.Flags().GetFloat64("mastery-points")
				body["mastery_points"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update outcome %s", outcomeID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/outcomes/"+outcomeID, body)
			if err != nil {
				return err
			}

			var outcome OutcomeSummary
			if err := json.Unmarshal(data, &outcome); err != nil {
				return fmt.Errorf("parse updated outcome: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(outcome)
			}
			fmt.Printf("Outcome %d updated.\n", outcome.ID)
			return nil
		},
	}

	cmd.Flags().String("outcome-id", "", "Canvas outcome ID (required)")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().Float64("mastery-points", 0, "New mastery points threshold")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newOutcomesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an outcome",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			outcomeID, _ := cmd.Flags().GetString("outcome-id")
			if outcomeID == "" {
				return fmt.Errorf("--outcome-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the outcome"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete outcome %s", outcomeID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/outcomes/"+outcomeID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "outcome_id": outcomeID})
			}
			fmt.Printf("Outcome %s deleted.\n", outcomeID)
			return nil
		},
	}

	cmd.Flags().String("outcome-id", "", "Canvas outcome ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newOutcomesGroupsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "List outcome groups for an account or course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			contextType, _ := cmd.Flags().GetString("context-type")
			contextID, _ := cmd.Flags().GetString("context-id")
			if contextID == "" {
				return fmt.Errorf("--context-id is required")
			}

			var path string
			switch strings.ToLower(contextType) {
			case "account":
				path = "/accounts/" + contextID + "/outcome_groups"
			case "course", "":
				path = "/courses/" + contextID + "/outcome_groups"
			default:
				return fmt.Errorf("--context-type must be Account or Course")
			}

			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var groups []map[string]any
			if err := json.Unmarshal(data, &groups); err != nil {
				return fmt.Errorf("parse outcome groups: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(groups)
			}

			if len(groups) == 0 {
				fmt.Println("No outcome groups found.")
				return nil
			}
			for _, g := range groups {
				id, _ := g["id"].(float64)
				title, _ := g["title"].(string)
				fmt.Printf("%-6.0f  %s\n", id, truncate(title, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("context-type", "Course", "Context type: Account or Course (required)")
	cmd.Flags().String("context-id", "", "ID of the account or course (required)")
	return cmd
}

func newOutcomesResultsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "results",
		Short: "List outcome results for a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}

			params := url.Values{}
			userIDs, _ := cmd.Flags().GetStringSlice("user-ids")
			for _, id := range userIDs {
				params.Add("user_ids[]", id)
			}
			outcomeIDs, _ := cmd.Flags().GetStringSlice("outcome-ids")
			for _, id := range outcomeIDs {
				params.Add("outcome_ids[]", id)
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/outcome_results", params)
			if err != nil {
				return err
			}

			var results map[string]any
			if err := json.Unmarshal(data, &results); err != nil {
				return fmt.Errorf("parse outcome results: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(results)
			}

			fmt.Printf("Outcome results for course %s retrieved. Use --json for full details.\n", courseID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().StringSlice("user-ids", nil, "Filter by user IDs (comma-separated)")
	cmd.Flags().StringSlice("outcome-ids", nil, "Filter by outcome IDs (comma-separated)")
	return cmd
}
