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
	cmd.AddCommand(newEnrollmentsCreateCmd(factory))
	cmd.AddCommand(newEnrollmentsDeactivateCmd(factory))
	cmd.AddCommand(newEnrollmentsReactivateCmd(factory))
	cmd.AddCommand(newEnrollmentsConcludeCmd(factory))
	cmd.AddCommand(newEnrollmentsDeleteCmd(factory))

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

func newEnrollmentsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Enroll a user in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			userID, _ := cmd.Flags().GetString("user-id")
			enrollmentType, _ := cmd.Flags().GetString("type")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}
			if enrollmentType == "" {
				return fmt.Errorf("--type is required")
			}

			enrollmentParams := map[string]any{
				"user_id": userID,
				"type":    enrollmentType,
			}
			if cmd.Flags().Changed("enrollment-state") {
				enrollmentState, _ := cmd.Flags().GetString("enrollment-state")
				enrollmentParams["enrollment_state"] = enrollmentState
			}
			body := map[string]any{"enrollment": enrollmentParams}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("enroll user %s as %s in course %s", userID, enrollmentType, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/enrollments", body)
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
			fmt.Printf("Enrollment created: %d  user:%d  type:%s  state:%s\n",
				enrollment.ID, enrollment.UserID, enrollment.Type, enrollment.EnrollmentState)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("user-id", "", "Canvas user ID to enroll (required)")
	cmd.Flags().String("type", "", "Enrollment type: StudentEnrollment|TeacherEnrollment|TaEnrollment|ObserverEnrollment|DesignerEnrollment (required)")
	cmd.Flags().String("enrollment-state", "", "Initial enrollment state: active|invited")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newEnrollmentsDeactivateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate an enrollment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			enrollmentID, _ := cmd.Flags().GetString("enrollment-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if enrollmentID == "" {
				return fmt.Errorf("--enrollment-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("deactivate enrollment %s in course %s", enrollmentID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/enrollments/" + enrollmentID + "?task=deactivate"
			data, err := client.Delete(ctx, path)
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
			fmt.Printf("Enrollment %s deactivated.\n", enrollmentID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("enrollment-id", "", "Canvas enrollment ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newEnrollmentsReactivateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reactivate",
		Short: "Reactivate an enrollment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			enrollmentID, _ := cmd.Flags().GetString("enrollment-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if enrollmentID == "" {
				return fmt.Errorf("--enrollment-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("reactivate enrollment %s in course %s", enrollmentID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/enrollments/" + enrollmentID + "/reactivate"
			data, err := client.Put(ctx, path, nil)
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
			fmt.Printf("Enrollment %s reactivated.\n", enrollmentID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("enrollment-id", "", "Canvas enrollment ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newEnrollmentsConcludeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conclude",
		Short: "Conclude an enrollment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			enrollmentID, _ := cmd.Flags().GetString("enrollment-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if enrollmentID == "" {
				return fmt.Errorf("--enrollment-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("conclude enrollment %s in course %s", enrollmentID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/enrollments/" + enrollmentID + "?task=conclude"
			data, err := client.Delete(ctx, path)
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
			fmt.Printf("Enrollment %s concluded.\n", enrollmentID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("enrollment-id", "", "Canvas enrollment ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newEnrollmentsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an enrollment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			enrollmentID, _ := cmd.Flags().GetString("enrollment-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if enrollmentID == "" {
				return fmt.Errorf("--enrollment-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the enrollment"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete enrollment %s in course %s", enrollmentID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/enrollments/" + enrollmentID + "?task=delete"
			_, err = client.Delete(ctx, path)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "enrollment_id": enrollmentID})
			}
			fmt.Printf("Enrollment %s deleted.\n", enrollmentID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("enrollment-id", "", "Canvas enrollment ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
