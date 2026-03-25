package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newRubricsCmd returns the parent "rubrics" command with all subcommands attached.
func newRubricsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rubrics",
		Short:   "Manage Canvas rubrics",
		Aliases: []string{"rubric"},
	}

	cmd.AddCommand(newRubricsListCmd(factory))
	cmd.AddCommand(newRubricsGetCmd(factory))

	return cmd
}

func newRubricsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rubrics for a course",
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

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/rubrics", params)
			if err != nil {
				return err
			}

			var rubrics []RubricSummary
			if err := json.Unmarshal(data, &rubrics); err != nil {
				return fmt.Errorf("parse rubrics: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(rubrics)
			}

			if len(rubrics) == 0 {
				fmt.Println("No rubrics found.")
				return nil
			}
			for _, r := range rubrics {
				fmt.Printf("%-6d  %-8.1f pts  %s\n", r.ID, r.PointsPossible, truncate(r.Title, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of rubrics to return")
	return cmd
}

func newRubricsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific rubric",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			rubricID, _ := cmd.Flags().GetString("rubric-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if rubricID == "" {
				return fmt.Errorf("--rubric-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/rubrics/"+rubricID, nil)
			if err != nil {
				return err
			}

			var rubric RubricSummary
			if err := json.Unmarshal(data, &rubric); err != nil {
				return fmt.Errorf("parse rubric: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(rubric)
			}

			fmt.Printf("ID:      %d\n", rubric.ID)
			fmt.Printf("Title:   %s\n", rubric.Title)
			fmt.Printf("Points:  %.1f\n", rubric.PointsPossible)
			if rubric.ReadOnly {
				fmt.Println("Read-Only: yes")
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("rubric-id", "", "Canvas rubric ID (required)")
	return cmd
}

