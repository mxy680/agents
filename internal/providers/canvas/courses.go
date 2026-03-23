package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newCoursesCmd returns the parent "courses" command with all subcommands attached.
func newCoursesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "courses",
		Short:   "Manage Canvas courses",
		Aliases: []string{"course"},
	}

	cmd.AddCommand(newCoursesListCmd(factory))
	cmd.AddCommand(newCoursesGetCmd(factory))

	return cmd
}

func newCoursesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List courses for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}
			params.Set("enrollment_state", "active")

			data, err := client.Get(ctx, "/courses", params)
			if err != nil {
				return err
			}

			var courses []CourseSummary
			if err := json.Unmarshal(data, &courses); err != nil {
				return fmt.Errorf("parse courses: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(courses)
			}

			if len(courses) == 0 {
				fmt.Println("No courses found.")
				return nil
			}
			for _, c := range courses {
				fmt.Printf("%-6d  %-10s  %s\n", c.ID, c.CourseCode, c.Name)
			}
			return nil
		},
	}

	cmd.Flags().Int("limit", 0, "Maximum number of courses to return")
	return cmd
}

func newCoursesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific course",
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

			data, err := client.Get(ctx, "/courses/"+courseID, nil)
			if err != nil {
				return err
			}

			var course CourseSummary
			if err := json.Unmarshal(data, &course); err != nil {
				return fmt.Errorf("parse course: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(course)
			}

			fmt.Printf("ID:           %d\n", course.ID)
			fmt.Printf("Name:         %s\n", course.Name)
			fmt.Printf("Code:         %s\n", course.CourseCode)
			fmt.Printf("State:        %s\n", course.WorkflowState)
			if course.StartAt != "" {
				fmt.Printf("Start:        %s\n", course.StartAt)
			}
			if course.EndAt != "" {
				fmt.Printf("End:          %s\n", course.EndAt)
			}
			if course.TotalStudents > 0 {
				fmt.Printf("Students:     %d\n", course.TotalStudents)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}
