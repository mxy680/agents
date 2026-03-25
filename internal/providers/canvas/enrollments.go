package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newEnrollmentsCmd returns the parent "enrollments" command with all subcommands attached.
func newEnrollmentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enrollments",
		Short:   "Manage Canvas course enrollments",
		Aliases: []string{"enroll"},
	}

	cmd.AddCommand(newEnrollmentsListCmd(factory))
	cmd.AddCommand(newEnrollmentsGetCmd(factory))

	return cmd
}

func newEnrollmentsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List enrollments for a course",
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

			enrollmentType, _ := cmd.Flags().GetString("type")
			state, _ := cmd.Flags().GetString("state")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if enrollmentType != "" {
				for _, t := range strings.Split(enrollmentType, ",") {
					params.Add("type[]", strings.TrimSpace(t))
				}
			}
			if state != "" {
				for _, s := range strings.Split(state, ",") {
					params.Add("state[]", strings.TrimSpace(s))
				}
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

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
				name := e.UserName
				if name == "" {
					name = strconv.Itoa(e.UserID)
				}
				fmt.Printf("%-8d  %-30s  %-25s  %s\n", e.ID, name, e.Type, e.EnrollmentState)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("type", "", "Filter by enrollment type: StudentEnrollment|TeacherEnrollment|TaEnrollment|ObserverEnrollment|DesignerEnrollment")
	cmd.Flags().String("state", "", "Filter by enrollment state: active|invited|completed|inactive|rejected")
	cmd.Flags().Int("limit", 0, "Maximum number of enrollments to return")
	return cmd
}

func newEnrollmentsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific enrollment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			enrollmentID, _ := cmd.Flags().GetString("enrollment-id")
			if enrollmentID == "" {
				return fmt.Errorf("--enrollment-id is required")
			}

			data, err := client.Get(ctx, "/accounts/self/enrollments/"+enrollmentID, nil)
			if err != nil {
				return err
			}

			var enrollment EnrollmentSummary
			if err := json.Unmarshal(data, &enrollment); err != nil {
				return fmt.Errorf("parse enrollment: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(enrollment)
			}

			fmt.Printf("ID:           %d\n", enrollment.ID)
			fmt.Printf("Course ID:    %d\n", enrollment.CourseID)
			fmt.Printf("User ID:      %d\n", enrollment.UserID)
			fmt.Printf("Type:         %s\n", enrollment.Type)
			fmt.Printf("State:        %s\n", enrollment.EnrollmentState)
			if enrollment.UserName != "" {
				fmt.Printf("User Name:    %s\n", enrollment.UserName)
			}
			if enrollment.CurrentGrade != "" {
				fmt.Printf("Grade:        %s (%.1f)\n", enrollment.CurrentGrade, enrollment.CurrentScore)
			}
			return nil
		},
	}

	cmd.Flags().String("enrollment-id", "", "Canvas enrollment ID (required)")
	return cmd
}
