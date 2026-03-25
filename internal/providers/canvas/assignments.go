package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newAssignmentsCmd returns the parent "assignments" command with all subcommands attached.
func newAssignmentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "assignments",
		Short:   "Manage Canvas assignments",
		Aliases: []string{"assignment", "assign"},
	}

	cmd.AddCommand(newAssignmentsListCmd(factory))
	cmd.AddCommand(newAssignmentsGetCmd(factory))

	return cmd
}

func newAssignmentsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List assignments for a course",
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

			data, err := client.Get(ctx, "/courses/"+courseID+"/assignments", params)
			if err != nil {
				return err
			}

			var assignments []AssignmentSummary
			if err := json.Unmarshal(data, &assignments); err != nil {
				return fmt.Errorf("parse assignments: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(assignments)
			}

			if len(assignments) == 0 {
				fmt.Println("No assignments found.")
				return nil
			}
			for _, a := range assignments {
				due := a.DueAt
				if due == "" {
					due = "no due date"
				}
				fmt.Printf("%-6d  %-8.1f pts  %-20s  %s\n", a.ID, a.PointsPossible, due, truncate(a.Name, 50))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of assignments to return")
	return cmd
}

func newAssignmentsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			assignmentID, _ := cmd.Flags().GetString("assignment-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if assignmentID == "" {
				return fmt.Errorf("--assignment-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/assignments/"+assignmentID, nil)
			if err != nil {
				return err
			}

			var assignment AssignmentSummary
			if err := json.Unmarshal(data, &assignment); err != nil {
				return fmt.Errorf("parse assignment: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(assignment)
			}

			fmt.Printf("ID:           %d\n", assignment.ID)
			fmt.Printf("Name:         %s\n", assignment.Name)
			fmt.Printf("Points:       %.1f\n", assignment.PointsPossible)
			fmt.Printf("Grading:      %s\n", assignment.GradingType)
			fmt.Printf("Published:    %v\n", assignment.Published)
			if assignment.DueAt != "" {
				fmt.Printf("Due:          %s\n", assignment.DueAt)
			}
			if assignment.Description != "" {
				fmt.Printf("Description:  %s\n", truncate(assignment.Description, 200))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	return cmd
}
