package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newGradesCmd returns the parent "grades" command with all subcommands attached.
func newGradesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grades",
		Short:   "View Canvas grade information",
		Aliases: []string{"grade"},
	}

	cmd.AddCommand(newGradesListCmd(factory))
	cmd.AddCommand(newGradesHistoryCmd(factory))

	return cmd
}

func newGradesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List enrollment grades for a course",
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
			params.Add("include[]", "grades")

			data, err := client.Get(ctx, "/courses/"+courseID+"/enrollments", params)
			if err != nil {
				return err
			}

			var enrollments []EnrollmentSummary
			if err := json.Unmarshal(data, &enrollments); err != nil {
				return fmt.Errorf("parse enrollments: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(enrollments)
			}

			if len(enrollments) == 0 {
				fmt.Println("No enrollments found.")
				return nil
			}
			for _, e := range enrollments {
				fmt.Printf("user:%-6d  current:%-6s (%-5.1f)  final:%-6s (%-5.1f)\n",
					e.UserID, e.CurrentGrade, e.CurrentScore, e.FinalGrade, e.FinalScore)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}

func newGradesHistoryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get grade change history for a course",
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

			assignmentID, _ := cmd.Flags().GetString("assignment-id")
			studentID, _ := cmd.Flags().GetString("student-id")

			params := url.Values{}
			if assignmentID != "" {
				params.Set("assignment_id", assignmentID)
			}
			if studentID != "" {
				params.Set("student_id", studentID)
			}

			data, err := client.Get(ctx, "/audit/grade_change/courses/"+courseID, params)
			if err != nil {
				return err
			}

			// Grade change events have a variable structure; always output as JSON.
			var result any
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse grade history: %w", err)
			}

			return cli.PrintJSON(result)
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Filter by assignment ID")
	cmd.Flags().String("student-id", "", "Filter by student ID")
	return cmd
}
