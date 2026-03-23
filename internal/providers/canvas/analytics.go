package canvas

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newAnalyticsCmd returns the parent "analytics" command with all subcommands attached.
func newAnalyticsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "analytics",
		Short:   "Access Canvas analytics data",
		Aliases: []string{"stats"},
	}

	cmd.AddCommand(newAnalyticsCourseCmd(factory))
	cmd.AddCommand(newAnalyticsAssignmentsCmd(factory))
	cmd.AddCommand(newAnalyticsStudentCmd(factory))
	cmd.AddCommand(newAnalyticsStudentAssignmentsCmd(factory))

	return cmd
}

func newAnalyticsCourseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "course",
		Short: "Get activity analytics for a course",
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

			data, err := client.Get(ctx, "/courses/"+courseID+"/analytics/activity", nil)
			if err != nil {
				return err
			}

			var result json.RawMessage
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse course analytics: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}

			fmt.Printf("Course analytics for course %s retrieved. Use --json for full details.\n", courseID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}

func newAnalyticsAssignmentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assignments",
		Short: "Get assignment analytics for a course",
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

			data, err := client.Get(ctx, "/courses/"+courseID+"/analytics/assignments", nil)
			if err != nil {
				return err
			}

			var results []map[string]any
			if err := json.Unmarshal(data, &results); err != nil {
				return fmt.Errorf("parse assignments analytics: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(results)
			}

			if len(results) == 0 {
				fmt.Println("No assignment analytics found.")
				return nil
			}
			for _, r := range results {
				title, _ := r["title"].(string)
				id, _ := r["assignment_id"].(float64)
				fmt.Printf("%-6.0f  %s\n", id, truncate(title, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}

func newAnalyticsStudentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "student",
		Short: "Get activity analytics for a student in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			studentID, _ := cmd.Flags().GetString("student-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if studentID == "" {
				return fmt.Errorf("--student-id is required")
			}

			path := "/courses/" + courseID + "/analytics/users/" + studentID + "/activity"
			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var result json.RawMessage
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse student analytics: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}

			fmt.Printf("Student %s analytics for course %s retrieved. Use --json for full details.\n", studentID, courseID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("student-id", "", "Canvas student user ID (required)")
	return cmd
}

func newAnalyticsStudentAssignmentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "student-assignments",
		Short: "Get assignment analytics for a student in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			studentID, _ := cmd.Flags().GetString("student-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if studentID == "" {
				return fmt.Errorf("--student-id is required")
			}

			path := "/courses/" + courseID + "/analytics/users/" + studentID + "/assignments"
			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var results []map[string]any
			if err := json.Unmarshal(data, &results); err != nil {
				return fmt.Errorf("parse student assignment analytics: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(results)
			}

			if len(results) == 0 {
				fmt.Println("No student assignment analytics found.")
				return nil
			}
			for _, r := range results {
				title, _ := r["title"].(string)
				id, _ := r["assignment_id"].(float64)
				score, _ := r["score"].(float64)
				fmt.Printf("%-6.0f  score=%-6.1f  %s\n", id, score, truncate(title, 50))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("student-id", "", "Canvas student user ID (required)")
	return cmd
}
